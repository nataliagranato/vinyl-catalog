package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/auth"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/observability"
)

// NewRouter creates and configures the Gin router with all routes and middlewares
func NewRouter(handler *VinylHandler, jwtSvc *auth.JWTService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("vinyl-catalog"))
	r.Use(observability.PrometheusMiddleware())

	// Health check
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Prometheus metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Auth routes (public)
	authGroup := r.Group("/api/v1/auth")
	authGroup.POST("/login", handler.Login)

	// Vinyl routes (protected)
	vinyls := r.Group("/api/v1/vinyls")
	vinyls.Use(JWTMiddleware(jwtSvc))
	{
		vinyls.GET("", handler.ListVinyls)
		vinyls.GET("/:id", handler.GetVinyl)
		vinyls.POST("", handler.CreateVinyl)
		vinyls.PUT("/:id", handler.UpdateVinyl)
		vinyls.DELETE("/:id", handler.DeleteVinyl)
		// Upload de capa (protegido)
		vinyls.POST("/:id/cover", handler.UploadCover)
		// Track routes
		vinyls.GET("/:id/tracks", handler.ListTracks)
		vinyls.POST("/:id/tracks", handler.CreateTrack)
		vinyls.PUT("/:id/tracks/:track_id", handler.UpdateTrack)
		vinyls.DELETE("/:id/tracks/:track_id", handler.DeleteTrack)
		// Favorites
		vinyls.POST("/:id/favorite", handler.ToggleFavorite)
	}

	// Profile — GET público, PUT e foto protegidos
	r.GET("/api/v1/profile", handler.GetProfile)

	profileRoutes := r.Group("/api/v1/profile")
	profileRoutes.Use(JWTMiddleware(jwtSvc))
	profileRoutes.PUT("", handler.UpdateProfile)
	profileRoutes.POST("/photo", handler.UploadProfilePhoto)

	// Servir arquivos de upload (público)
	r.Static("/uploads", "./uploads")

	return r
}

// JWTMiddleware validates Bearer tokens and sets "username" in context.
// Accepts both "Bearer <token>" and just "<token>" for Swagger UI compatibility.
func JWTMiddleware(jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "authorization header required"})
			return
		}
		tokenStr := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}
		claims, err := jwtSvc.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
			return
		}
		c.Set("username", claims.Username)
		c.Next()
	}
}
