// @title           Vinyl Catalog API
// @version         1.0
// @description     API REST para catalogação de discos de vinil
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Autenticação JWT. Faça POST em /auth/login e cole apenas o valor do campo "token" no campo Authorize abaixo (sem precisar digitar "Bearer ").

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/nataliagranato/vinyl-catalog/docs/swagger"
	pgAdapter "github.com/nataliagranato/vinyl-catalog/internal/adapters/secondary/postgres"
	httpAdapter "github.com/nataliagranato/vinyl-catalog/internal/adapters/primary/http"
	"github.com/nataliagranato/vinyl-catalog/internal/application"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/auth"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/config"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/database"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/observability"
)

func main() {
	// Structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load .env if present
	if err := godotenv.Load(); err != nil {
		slog.Debug("no .env file found, using environment variables")
	}

	cfg := config.Load()

	// Tracing setup
	ctx := context.Background()
	tp, err := observability.NewTracerProvider(ctx, cfg.OTELEndpoint, cfg.OTELServiceName, cfg.ServiceVersion, cfg.AppEnv)
	if err != nil {
		slog.Warn("tracing unavailable, continuing without traces", "error", err)
	} else {
		defer func() {
			if err := tp.Shutdown(ctx); err != nil {
				slog.Error("failed to shutdown tracer provider", "error", err)
			}
		}()
	}

	// Database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Auto-migrate
	if err := pgAdapter.AutoMigrateAll(db); err != nil {
		slog.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}

	// Dependency wiring
	repo := pgAdapter.NewVinylRepository(db)
	trackRepo := pgAdapter.NewTrackRepository(db)
	profileRepo := pgAdapter.NewProfileRepository(db)
	svc := application.NewVinylService(repo)
	trackSvc := application.NewTrackService(trackRepo)
	jwtSvc := auth.NewJWTServiceFromConfig(cfg.JWTSecret, cfg.JWTExpirationHours)
	handler := httpAdapter.NewVinylHandlerWithAuth(svc, trackSvc, profileRepo, jwtSvc, cfg.AdminUsername, cfg.AdminPassword)
	router := httpAdapter.NewRouter(handler, jwtSvc)

	// HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.AppPort, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}
	slog.Info("server stopped")
}
