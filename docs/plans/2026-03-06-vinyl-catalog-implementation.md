# Vinyl Catalog Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Construir um sistema de catalogação de discos de vinil com API REST em Go, usando Clean Architecture e Arquitetura Hexagonal, com PostgreSQL, JWT, Swagger e observabilidade completa.

**Architecture:** Hexagonal (Ports & Adapters) com Clean Architecture. O domínio define interfaces puras; adapters primários (HTTP/Gin) e secundários (PostgreSQL/GORM) implementam essas interfaces. Dependências sempre apontam para dentro: adapters → application → domain.

**Tech Stack:** Go 1.22+, Gin, GORM, PostgreSQL, golang-jwt/jwt, swaggo/swag, Prometheus, OpenTelemetry, Jaeger, slog, Docker, mise.

---

## Task 1: Scaffolding do Projeto

**Files:**
- Create: `go.mod`
- Create: `.env.example`
- Create: `mise.toml`
- Create: `.gitignore`

**Step 1: Inicializar módulo Go**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
go mod init github.com/nataliagranato/vinyl-catalog
```

Expected: `go.mod` criado com `module github.com/nataliagranato/vinyl-catalog` e `go 1.22`

**Step 2: Instalar dependências**

```bash
go get github.com/gin-gonic/gin@latest
go get gorm.io/gorm@latest
go get gorm.io/driver/postgres@latest
go get github.com/golang-jwt/jwt/v5@latest
go get github.com/google/uuid@latest
go get github.com/prometheus/client_golang@latest
go get go.opentelemetry.io/otel@latest
go get go.opentelemetry.io/otel/sdk@latest
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@latest
go get go.opentelemetry.io/otel/trace@latest
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin@latest
go get github.com/swaggo/swag@latest
go get github.com/swaggo/gin-swagger@latest
go get github.com/swaggo/files@latest
go get github.com/joho/godotenv@latest
```

**Step 3: Criar `.env.example`**

```env
# App
APP_PORT=8080
APP_ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=vinyl_catalog
DB_SSLMODE=disable

# Auth
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRATION_HOURS=24
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
OTEL_SERVICE_NAME=vinyl-catalog
```

**Step 4: Criar `mise.toml`**

```toml
[tools]
go = "1.22"

[tasks.dev]
description = "Run app with hot reload"
run = "go run ./cmd/api/main.go"

[tasks.build]
description = "Build binary"
run = "go build -o bin/vinyl-catalog ./cmd/api/main.go"

[tasks.test]
description = "Run all tests"
run = "go test ./... -v -cover"

[tasks.swag]
description = "Generate Swagger docs"
run = "swag init -g cmd/api/main.go -o docs/swagger"

[tasks.lint]
description = "Run linter"
run = "golangci-lint run ./..."

[tasks."docker:up"]
description = "Start all services"
run = "docker compose up -d --build"

[tasks."docker:down"]
description = "Stop all services"
run = "docker compose down"

[tasks."docker:logs"]
description = "Follow app logs"
run = "docker compose logs -f app"
```

**Step 5: Criar `.gitignore`**

```gitignore
# Binaries
bin/
*.exe

# Environment
.env
.envrc

# Docs gerados
docs/swagger/

# Go
vendor/
*.test
*.out

# IDE
.idea/
.vscode/
*.swp
```

**Step 6: Commit**

```bash
git init
git add go.mod go.sum .env.example mise.toml .gitignore
git commit -m "chore: initialize project scaffold with Go modules and mise"
```

---

## Task 2: Camada de Domínio

**Files:**
- Create: `internal/domain/vinyl.go`
- Create: `internal/domain/ports.go`
- Create: `internal/domain/errors.go`
- Test: `internal/domain/vinyl_test.go`

**Step 1: Escrever o teste da entidade**

```go
// internal/domain/vinyl_test.go
package domain_test

import (
	"testing"
	"time"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

func TestVinyl_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		vinyl   domain.Vinyl
		wantErr bool
	}{
		{
			name: "valid vinyl",
			vinyl: domain.Vinyl{
				Title:  "Kind of Blue",
				Artist: "Miles Davis",
				Year:   1959,
				Genre:  "Jazz",
				Label:  "Columbia",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			vinyl: domain.Vinyl{
				Artist: "Miles Davis",
				Year:   1959,
				Genre:  "Jazz",
				Label:  "Columbia",
			},
			wantErr: true,
		},
		{
			name: "missing artist",
			vinyl: domain.Vinyl{
				Title: "Kind of Blue",
				Year:  1959,
				Genre: "Jazz",
				Label: "Columbia",
			},
			wantErr: true,
		},
		{
			name: "invalid year",
			vinyl: domain.Vinyl{
				Title:  "Kind of Blue",
				Artist: "Miles Davis",
				Year:   1800,
				Genre:  "Jazz",
				Label:  "Columbia",
			},
			wantErr: true,
		},
		{
			name: "future year",
			vinyl: domain.Vinyl{
				Title:  "Future Album",
				Artist: "Artist",
				Year:   time.Now().Year() + 2,
				Genre:  "Pop",
				Label:  "Label",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vinyl.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

**Step 2: Rodar o teste para verificar que falha**

```bash
go test ./internal/domain/... -v
```

Expected: FAIL — `domain` package not found

**Step 3: Criar `internal/domain/errors.go`**

```go
package domain

import "errors"

var (
	ErrVinylNotFound    = errors.New("vinyl not found")
	ErrVinylTitleEmpty  = errors.New("vinyl title is required")
	ErrVinylArtistEmpty = errors.New("vinyl artist is required")
	ErrVinylYearInvalid = errors.New("vinyl year must be between 1860 and current year")
)
```

**Step 4: Criar `internal/domain/vinyl.go`**

```go
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Vinyl struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Artist    string    `json:"artist"`
	Year      int       `json:"year"`
	Genre     string    `json:"genre"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewVinyl(title, artist string, year int, genre, label string) (*Vinyl, error) {
	v := &Vinyl{
		ID:     uuid.NewString(),
		Title:  title,
		Artist: artist,
		Year:   year,
		Genre:  genre,
		Label:  label,
	}
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *Vinyl) Validate() error {
	if v.Title == "" {
		return ErrVinylTitleEmpty
	}
	if v.Artist == "" {
		return ErrVinylArtistEmpty
	}
	currentYear := time.Now().Year()
	if v.Year < 1860 || v.Year > currentYear+1 {
		return fmt.Errorf("%w: got %d", ErrVinylYearInvalid, v.Year)
	}
	return nil
}
```

**Step 5: Criar `internal/domain/ports.go`**

```go
package domain

import "context"

type VinylRepository interface {
	Create(ctx context.Context, v *Vinyl) error
	FindByID(ctx context.Context, id string) (*Vinyl, error)
	FindAll(ctx context.Context) ([]Vinyl, error)
	Update(ctx context.Context, v *Vinyl) error
	Delete(ctx context.Context, id string) error
}

type VinylService interface {
	CreateVinyl(ctx context.Context, v *Vinyl) error
	GetVinyl(ctx context.Context, id string) (*Vinyl, error)
	ListVinyls(ctx context.Context) ([]Vinyl, error)
	UpdateVinyl(ctx context.Context, v *Vinyl) error
	DeleteVinyl(ctx context.Context, id string) error
}
```

**Step 6: Rodar testes e verificar que passam**

```bash
go test ./internal/domain/... -v
```

Expected: PASS — todos os 5 casos de teste

**Step 7: Commit**

```bash
git add internal/domain/
git commit -m "feat: add domain layer with Vinyl entity, ports and validation"
```

---

## Task 3: Configuração de Infraestrutura

**Files:**
- Create: `internal/infrastructure/config/config.go`
- Test: `internal/infrastructure/config/config_test.go`

**Step 1: Escrever o teste**

```go
// internal/infrastructure/config/config_test.go
package config_test

import (
	"os"
	"testing"

	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/config"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()
	cfg := config.Load()

	if cfg.AppPort != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.AppPort)
	}
	if cfg.DBSSLMode != "disable" {
		t.Errorf("expected default sslmode disable, got %s", cfg.DBSSLMode)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_HOST", "myhost")
	defer os.Clearenv()

	cfg := config.Load()

	if cfg.AppPort != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.AppPort)
	}
	if cfg.DBHost != "myhost" {
		t.Errorf("expected host myhost, got %s", cfg.DBHost)
	}
}
```

**Step 2: Rodar para verificar falha**

```bash
go test ./internal/infrastructure/config/... -v
```

Expected: FAIL

**Step 3: Criar `internal/infrastructure/config/config.go`**

```go
package config

import "os"

type Config struct {
	AppPort  string
	AppEnv   string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret           string
	JWTExpirationHours  string
	AdminUsername       string
	AdminPassword       string

	OTELEndpoint    string
	OTELServiceName string
}

func Load() *Config {
	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "vinyl_catalog"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:          getEnv("JWT_SECRET", "change-me"),
		JWTExpirationHours: getEnv("JWT_EXPIRATION_HOURS", "24"),
		AdminUsername:      getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:      getEnv("ADMIN_PASSWORD", "admin"),

		OTELEndpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		OTELServiceName: getEnv("OTEL_SERVICE_NAME", "vinyl-catalog"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

**Step 4: Rodar testes**

```bash
go test ./internal/infrastructure/config/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/infrastructure/config/
git commit -m "feat: add config infrastructure with env var loading"
```

---

## Task 4: Infraestrutura de Banco de Dados

**Files:**
- Create: `internal/infrastructure/database/postgres.go`

**Step 1: Criar `internal/infrastructure/database/postgres.go`**

```go
package database

import (
	"fmt"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/config"
)

func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	logLevel := logger.Silent
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("database connected", "host", cfg.DBHost, "name", cfg.DBName)
	return db, nil
}
```

**Step 2: Commit**

```bash
git add internal/infrastructure/database/
git commit -m "feat: add PostgreSQL database infrastructure"
```

---

## Task 5: Infraestrutura de Autenticação JWT

**Files:**
- Create: `internal/infrastructure/auth/jwt.go`
- Test: `internal/infrastructure/auth/jwt_test.go`

**Step 1: Escrever o teste**

```go
// internal/infrastructure/auth/jwt_test.go
package auth_test

import (
	"testing"
	"time"

	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/auth"
)

func TestJWT_GenerateAndValidate(t *testing.T) {
	svc := auth.NewJWTService("test-secret", 1)

	token, err := svc.GenerateToken("admin")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken() returned empty token")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.Username != "admin" {
		t.Errorf("expected username admin, got %s", claims.Username)
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	svc := auth.NewJWTService("test-secret", 1)
	_, err := svc.ValidateToken("invalid.token.here")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestJWT_WrongSecret(t *testing.T) {
	svc1 := auth.NewJWTService("secret-a", 1)
	svc2 := auth.NewJWTService("secret-b", 1)

	token, _ := svc1.GenerateToken("admin")
	_, err := svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for token signed with different secret")
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	// expiration de 0 horas = token imediatamente expirado
	svc := auth.NewJWTService("test-secret", -1)
	token, _ := svc.GenerateToken("admin")
	time.Sleep(10 * time.Millisecond)
	_, err := svc.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}
```

**Step 2: Rodar para verificar falha**

```bash
go test ./internal/infrastructure/auth/... -v
```

Expected: FAIL

**Step 3: Criar `internal/infrastructure/auth/jwt.go`**

```go
package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret     []byte
	expiration time.Duration
}

func NewJWTService(secret string, expirationHours int) *JWTService {
	return &JWTService{
		secret:     []byte(secret),
		expiration: time.Duration(expirationHours) * time.Hour,
	}
}

func NewJWTServiceFromConfig(secret, expirationHours string) *JWTService {
	hours, err := strconv.Atoi(expirationHours)
	if err != nil {
		hours = 24
	}
	return NewJWTService(secret, hours)
}

func (s *JWTService) GenerateToken(username string) (string, error) {
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}
```

**Step 4: Rodar testes**

```bash
go test ./internal/infrastructure/auth/... -v
```

Expected: PASS — 4 testes

**Step 5: Commit**

```bash
git add internal/infrastructure/auth/
git commit -m "feat: add JWT service for token generation and validation"
```

---

## Task 6: Observabilidade — Métricas Prometheus

**Files:**
- Create: `internal/infrastructure/observability/metrics.go`

**Step 1: Criar `internal/infrastructure/observability/metrics.go`**

```go
package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
```

**Step 2: Commit**

```bash
git add internal/infrastructure/observability/metrics.go
git commit -m "feat: add Prometheus metrics middleware"
```

---

## Task 7: Observabilidade — Tracing OpenTelemetry

**Files:**
- Create: `internal/infrastructure/observability/tracing.go`

**Step 1: Criar `internal/infrastructure/observability/tracing.go`**

```go
package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewTracerProvider(ctx context.Context, endpoint, serviceName string) (*sdktrace.TracerProvider, error) {
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}
```

**Step 2: Commit**

```bash
git add internal/infrastructure/observability/tracing.go
git commit -m "feat: add OpenTelemetry tracing with OTLP gRPC exporter"
```

---

## Task 8: Adapter Secundário — Repositório PostgreSQL

**Files:**
- Create: `internal/adapters/secondary/postgres/vinyl_repo.go`
- Test: `internal/adapters/secondary/postgres/vinyl_repo_test.go`

**Step 1: Criar o model GORM separado da entidade de domínio**

```go
// internal/adapters/secondary/postgres/vinyl_repo.go
package postgres

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"gorm.io/gorm"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

type vinylModel struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	Title     string    `gorm:"not null"`
	Artist    string    `gorm:"not null"`
	Year      int       `gorm:"not null"`
	Genre     string
	Label     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (vinylModel) TableName() string { return "vinyls" }

func toModel(v *domain.Vinyl) *vinylModel {
	return &vinylModel{
		ID:        v.ID,
		Title:     v.Title,
		Artist:    v.Artist,
		Year:      v.Year,
		Genre:     v.Genre,
		Label:     v.Label,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

func toDomain(m *vinylModel) *domain.Vinyl {
	return &domain.Vinyl{
		ID:        m.ID,
		Title:     m.Title,
		Artist:    m.Artist,
		Year:      m.Year,
		Genre:     m.Genre,
		Label:     m.Label,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type VinylRepository struct {
	db *gorm.DB
}

func NewVinylRepository(db *gorm.DB) *VinylRepository {
	return &VinylRepository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&vinylModel{})
}

func (r *VinylRepository) Create(ctx context.Context, v *domain.Vinyl) error {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "VinylRepository.Create")
	defer span.End()

	return r.db.WithContext(ctx).Create(toModel(v)).Error
}

func (r *VinylRepository) FindByID(ctx context.Context, id string) (*domain.Vinyl, error) {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "VinylRepository.FindByID")
	defer span.End()

	var m vinylModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrVinylNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(&m), nil
}

func (r *VinylRepository) FindAll(ctx context.Context) ([]domain.Vinyl, error) {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "VinylRepository.FindAll")
	defer span.End()

	var models []vinylModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}
	vinyls := make([]domain.Vinyl, len(models))
	for i, m := range models {
		v := toDomain(&m)
		vinyls[i] = *v
	}
	return vinyls, nil
}

func (r *VinylRepository) Update(ctx context.Context, v *domain.Vinyl) error {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "VinylRepository.Update")
	defer span.End()

	result := r.db.WithContext(ctx).Model(&vinylModel{}).Where("id = ?", v.ID).Updates(toModel(v))
	if result.RowsAffected == 0 {
		return domain.ErrVinylNotFound
	}
	return result.Error
}

func (r *VinylRepository) Delete(ctx context.Context, id string) error {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "VinylRepository.Delete")
	defer span.End()

	result := r.db.WithContext(ctx).Delete(&vinylModel{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return domain.ErrVinylNotFound
	}
	return result.Error
}
```

**Step 2: Escrever teste unitário com mock do DB**

```go
// internal/adapters/secondary/postgres/vinyl_repo_test.go
package postgres_test

import (
	"context"
	"testing"

	"github.com/nataliagranato/vinyl-catalog/internal/adapters/secondary/postgres"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Nota: este teste requer um PostgreSQL rodando.
// Para CI/CD, use testcontainers-go ou um banco de teste dedicado.
// Aqui validamos apenas a compilação e interface correta.

func TestVinylRepository_ImplementsInterface(t *testing.T) {
	// Verifica em tempo de compilação que VinylRepository implementa VinylRepository interface
	var _ domain.VinylRepository = (*postgres.VinylRepository)(nil)
}

func TestVinylRepository_NotFound(t *testing.T) {
	// Este teste requer DB real — pule se não houver conexão
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=vinyl_catalog_test sslmode=disable"
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
	}

	if err := postgres.AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate error: %v", err)
	}

	repo := postgres.NewVinylRepository(db)
	_, err = repo.FindByID(context.Background(), "non-existent-id")
	if err != domain.ErrVinylNotFound {
		t.Errorf("expected ErrVinylNotFound, got %v", err)
	}
}
```

**Step 3: Rodar testes**

```bash
go test ./internal/adapters/secondary/postgres/... -v
```

Expected: PASS (compilação ok, test de integração pulado se não houver DB)

**Step 4: Commit**

```bash
git add internal/adapters/secondary/
git commit -m "feat: add PostgreSQL vinyl repository adapter with GORM and OTel tracing"
```

---

## Task 9: Camada de Aplicação — Vinyl Service

**Files:**
- Create: `internal/application/vinyl_service.go`
- Test: `internal/application/vinyl_service_test.go`

**Step 1: Escrever o teste com mock do repositório**

```go
// internal/application/vinyl_service_test.go
package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nataliagranato/vinyl-catalog/internal/application"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

// mockRepo implementa domain.VinylRepository para testes
type mockRepo struct {
	vinyls map[string]*domain.Vinyl
	err    error
}

func newMockRepo() *mockRepo {
	return &mockRepo{vinyls: make(map[string]*domain.Vinyl)}
}

func (m *mockRepo) Create(_ context.Context, v *domain.Vinyl) error {
	if m.err != nil {
		return m.err
	}
	m.vinyls[v.ID] = v
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id string) (*domain.Vinyl, error) {
	if m.err != nil {
		return nil, m.err
	}
	v, ok := m.vinyls[id]
	if !ok {
		return nil, domain.ErrVinylNotFound
	}
	return v, nil
}

func (m *mockRepo) FindAll(_ context.Context) ([]domain.Vinyl, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]domain.Vinyl, 0, len(m.vinyls))
	for _, v := range m.vinyls {
		result = append(result, *v)
	}
	return result, nil
}

func (m *mockRepo) Update(_ context.Context, v *domain.Vinyl) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.vinyls[v.ID]; !ok {
		return domain.ErrVinylNotFound
	}
	m.vinyls[v.ID] = v
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.vinyls[id]; !ok {
		return domain.ErrVinylNotFound
	}
	delete(m.vinyls, id)
	return nil
}

func TestVinylService_CreateAndGet(t *testing.T) {
	repo := newMockRepo()
	svc := application.NewVinylService(repo)
	ctx := context.Background()

	vinyl := &domain.Vinyl{
		ID:     "test-id",
		Title:  "Kind of Blue",
		Artist: "Miles Davis",
		Year:   1959,
		Genre:  "Jazz",
		Label:  "Columbia",
	}

	if err := svc.CreateVinyl(ctx, vinyl); err != nil {
		t.Fatalf("CreateVinyl() error = %v", err)
	}

	got, err := svc.GetVinyl(ctx, "test-id")
	if err != nil {
		t.Fatalf("GetVinyl() error = %v", err)
	}
	if got.Title != "Kind of Blue" {
		t.Errorf("expected title 'Kind of Blue', got %s", got.Title)
	}
}

func TestVinylService_GetNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := application.NewVinylService(repo)

	_, err := svc.GetVinyl(context.Background(), "not-found")
	if !errors.Is(err, domain.ErrVinylNotFound) {
		t.Errorf("expected ErrVinylNotFound, got %v", err)
	}
}

func TestVinylService_UpdateAndDelete(t *testing.T) {
	repo := newMockRepo()
	svc := application.NewVinylService(repo)
	ctx := context.Background()

	vinyl := &domain.Vinyl{ID: "id-1", Title: "Original", Artist: "Artist", Year: 2000, Genre: "Rock", Label: "Label"}
	_ = svc.CreateVinyl(ctx, vinyl)

	vinyl.Title = "Updated"
	if err := svc.UpdateVinyl(ctx, vinyl); err != nil {
		t.Fatalf("UpdateVinyl() error = %v", err)
	}

	if err := svc.DeleteVinyl(ctx, "id-1"); err != nil {
		t.Fatalf("DeleteVinyl() error = %v", err)
	}

	_, err := svc.GetVinyl(ctx, "id-1")
	if !errors.Is(err, domain.ErrVinylNotFound) {
		t.Errorf("expected ErrVinylNotFound after delete, got %v", err)
	}
}

func TestVinylService_ListVinyls(t *testing.T) {
	repo := newMockRepo()
	svc := application.NewVinylService(repo)
	ctx := context.Background()

	for i, title := range []string{"Album A", "Album B", "Album C"} {
		v := &domain.Vinyl{ID: fmt.Sprintf("id-%d", i), Title: title, Artist: "Artist", Year: 2000, Genre: "Pop", Label: "Label"}
		_ = svc.CreateVinyl(ctx, v)
	}

	vinyls, err := svc.ListVinyls(ctx)
	if err != nil {
		t.Fatalf("ListVinyls() error = %v", err)
	}
	if len(vinyls) != 3 {
		t.Errorf("expected 3 vinyls, got %d", len(vinyls))
	}
}
```

Note: adicionar `"fmt"` nos imports do teste.

**Step 2: Rodar para verificar falha**

```bash
go test ./internal/application/... -v
```

Expected: FAIL

**Step 3: Criar `internal/application/vinyl_service.go`**

```go
package application

import (
	"context"

	"go.opentelemetry.io/otel"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

type vinylService struct {
	repo domain.VinylRepository
}

func NewVinylService(repo domain.VinylRepository) domain.VinylService {
	return &vinylService{repo: repo}
}

func (s *vinylService) CreateVinyl(ctx context.Context, v *domain.Vinyl) error {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "VinylService.CreateVinyl")
	defer span.End()
	return s.repo.Create(ctx, v)
}

func (s *vinylService) GetVinyl(ctx context.Context, id string) (*domain.Vinyl, error) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "VinylService.GetVinyl")
	defer span.End()
	return s.repo.FindByID(ctx, id)
}

func (s *vinylService) ListVinyls(ctx context.Context) ([]domain.Vinyl, error) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "VinylService.ListVinyls")
	defer span.End()
	return s.repo.FindAll(ctx)
}

func (s *vinylService) UpdateVinyl(ctx context.Context, v *domain.Vinyl) error {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "VinylService.UpdateVinyl")
	defer span.End()
	return s.repo.Update(ctx, v)
}

func (s *vinylService) DeleteVinyl(ctx context.Context, id string) error {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "VinylService.DeleteVinyl")
	defer span.End()
	return s.repo.Delete(ctx, id)
}
```

**Step 4: Rodar testes**

```bash
go test ./internal/application/... -v
```

Expected: PASS — 4 testes

**Step 5: Commit**

```bash
git add internal/application/
git commit -m "feat: add application layer with VinylService use cases"
```

---

## Task 10: Adapter Primário — DTOs HTTP

**Files:**
- Create: `internal/adapters/primary/http/dto.go`

**Step 1: Criar `internal/adapters/primary/http/dto.go`**

```go
package http

import "github.com/nataliagranato/vinyl-catalog/internal/domain"

// CreateVinylRequest representa o body de criação de um disco
// @Description Request para criar um novo disco de vinil
type CreateVinylRequest struct {
	Title  string `json:"title"  binding:"required" example:"Kind of Blue"`
	Artist string `json:"artist" binding:"required" example:"Miles Davis"`
	Year   int    `json:"year"   binding:"required,min=1860" example:"1959"`
	Genre  string `json:"genre"  example:"Jazz"`
	Label  string `json:"label"  example:"Columbia"`
}

// UpdateVinylRequest representa o body de atualização de um disco
type UpdateVinylRequest struct {
	Title  string `json:"title"  binding:"required" example:"Kind of Blue"`
	Artist string `json:"artist" binding:"required" example:"Miles Davis"`
	Year   int    `json:"year"   binding:"required,min=1860" example:"1959"`
	Genre  string `json:"genre"  example:"Jazz"`
	Label  string `json:"label"  example:"Columbia"`
}

// LoginRequest representa as credenciais de login
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"admin"`
}

// TokenResponse retorna o JWT gerado
type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Error string `json:"error" example:"vinyl not found"`
}

// VinylResponse representa um disco na resposta
type VinylResponse struct {
	ID        string `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	Title     string `json:"title"      example:"Kind of Blue"`
	Artist    string `json:"artist"     example:"Miles Davis"`
	Year      int    `json:"year"       example:"1959"`
	Genre     string `json:"genre"      example:"Jazz"`
	Label     string `json:"label"      example:"Columbia"`
	CreatedAt string `json:"created_at" example:"2026-03-06T00:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2026-03-06T00:00:00Z"`
}

func toVinylResponse(v *domain.Vinyl) VinylResponse {
	return VinylResponse{
		ID:        v.ID,
		Title:     v.Title,
		Artist:    v.Artist,
		Year:      v.Year,
		Genre:     v.Genre,
		Label:     v.Label,
		CreatedAt: v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: v.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
```

**Step 2: Commit**

```bash
git add internal/adapters/primary/http/dto.go
git commit -m "feat: add HTTP DTOs for vinyl API"
```

---

## Task 11: Adapter Primário — Handlers HTTP

**Files:**
- Create: `internal/adapters/primary/http/handler.go`
- Test: `internal/adapters/primary/http/handler_test.go`

**Step 1: Escrever testes dos handlers**

```go
// internal/adapters/primary/http/handler_test.go
package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
	httpAdapter "github.com/nataliagranato/vinyl-catalog/internal/adapters/primary/http"
)

type mockService struct {
	vinyls []domain.Vinyl
	err    error
}

func (m *mockService) CreateVinyl(_ context.Context, v *domain.Vinyl) error { return m.err }
func (m *mockService) GetVinyl(_ context.Context, id string) (*domain.Vinyl, error) {
	if m.err != nil { return nil, m.err }
	for _, v := range m.vinyls {
		if v.ID == id { return &v, nil }
	}
	return nil, domain.ErrVinylNotFound
}
func (m *mockService) ListVinyls(_ context.Context) ([]domain.Vinyl, error) {
	return m.vinyls, m.err
}
func (m *mockService) UpdateVinyl(_ context.Context, v *domain.Vinyl) error { return m.err }
func (m *mockService) DeleteVinyl(_ context.Context, id string) error       { return m.err }

func setupRouter(svc domain.VinylService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httpAdapter.NewVinylHandler(svc)
	v1 := r.Group("/api/v1/vinyls")
	v1.GET("", h.ListVinyls)
	v1.GET("/:id", h.GetVinyl)
	v1.POST("", h.CreateVinyl)
	v1.PUT("/:id", h.UpdateVinyl)
	v1.DELETE("/:id", h.DeleteVinyl)
	return r
}

func TestHandler_ListVinyls(t *testing.T) {
	svc := &mockService{vinyls: []domain.Vinyl{
		{ID: "1", Title: "Album A", Artist: "Artist", Year: 2000, Genre: "Rock", Label: "Label"},
	}}
	r := setupRouter(svc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandler_GetVinyl_NotFound(t *testing.T) {
	svc := &mockService{}
	r := setupRouter(svc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls/not-found", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandler_CreateVinyl(t *testing.T) {
	svc := &mockService{}
	r := setupRouter(svc)
	body, _ := json.Marshal(map[string]any{
		"title": "Kind of Blue", "artist": "Miles Davis", "year": 1959,
		"genre": "Jazz", "label": "Columbia",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/vinyls", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestHandler_CreateVinyl_InvalidBody(t *testing.T) {
	svc := &mockService{}
	r := setupRouter(svc)
	body := bytes.NewBufferString(`{"title": ""}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/vinyls", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
```

**Step 2: Rodar para verificar falha**

```bash
go test ./internal/adapters/primary/http/... -v
```

Expected: FAIL

**Step 3: Criar `internal/adapters/primary/http/handler.go`**

```go
package http

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/auth"
)

type VinylHandler struct {
	service    domain.VinylService
	jwtService *auth.JWTService
	adminUser  string
	adminPass  string
}

func NewVinylHandler(service domain.VinylService) *VinylHandler {
	return &VinylHandler{service: service}
}

func NewVinylHandlerWithAuth(service domain.VinylService, jwtSvc *auth.JWTService, user, pass string) *VinylHandler {
	return &VinylHandler{service: service, jwtService: jwtSvc, adminUser: user, adminPass: pass}
}

// Login godoc
// @Summary      Autenticar usuário
// @Description  Gera um token JWT para acesso à API
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Credenciais"
// @Success      200 {object} TokenResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /auth/login [post]
func (h *VinylHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if req.Username != h.adminUser || req.Password != h.adminPass {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
		return
	}
	token, err := h.jwtService.GenerateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, TokenResponse{Token: token})
}

// ListVinyls godoc
// @Summary      Listar discos
// @Description  Retorna todos os discos de vinil cadastrados
// @Tags         vinyls
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array}  VinylResponse
// @Failure      500 {object} ErrorResponse
// @Router       /vinyls [get]
func (h *VinylHandler) ListVinyls(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.ListVinyls")
	defer span.End()

	vinyls, err := h.service.ListVinyls(ctx)
	if err != nil {
		slog.Error("failed to list vinyls", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	resp := make([]VinylResponse, len(vinyls))
	for i := range vinyls {
		resp[i] = toVinylResponse(&vinyls[i])
	}
	c.JSON(http.StatusOK, resp)
}

// GetVinyl godoc
// @Summary      Buscar disco por ID
// @Description  Retorna um disco de vinil pelo seu ID
// @Tags         vinyls
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Vinyl ID"
// @Success      200 {object} VinylResponse
// @Failure      404 {object} ErrorResponse
// @Router       /vinyls/{id} [get]
func (h *VinylHandler) GetVinyl(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.GetVinyl")
	defer span.End()

	id := c.Param("id")
	vinyl, err := h.service.GetVinyl(ctx, id)
	if errors.Is(err, domain.ErrVinylNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "vinyl not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, toVinylResponse(vinyl))
}

// CreateVinyl godoc
// @Summary      Criar disco
// @Description  Cadastra um novo disco de vinil
// @Tags         vinyls
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateVinylRequest true "Dados do disco"
// @Success      201 {object} VinylResponse
// @Failure      400 {object} ErrorResponse
// @Router       /vinyls [post]
func (h *VinylHandler) CreateVinyl(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.CreateVinyl")
	defer span.End()

	var req CreateVinylRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	vinyl := &domain.Vinyl{
		ID:        uuid.NewString(),
		Title:     req.Title,
		Artist:    req.Artist,
		Year:      req.Year,
		Genre:     req.Genre,
		Label:     req.Label,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := vinyl.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.service.CreateVinyl(ctx, vinyl); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, toVinylResponse(vinyl))
}

// UpdateVinyl godoc
// @Summary      Atualizar disco
// @Description  Atualiza os dados de um disco de vinil existente
// @Tags         vinyls
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path string             true "Vinyl ID"
// @Param        request body UpdateVinylRequest true "Dados atualizados"
// @Success      200 {object} VinylResponse
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /vinyls/{id} [put]
func (h *VinylHandler) UpdateVinyl(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.UpdateVinyl")
	defer span.End()

	id := c.Param("id")
	var req UpdateVinylRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	vinyl := &domain.Vinyl{
		ID:        id,
		Title:     req.Title,
		Artist:    req.Artist,
		Year:      req.Year,
		Genre:     req.Genre,
		Label:     req.Label,
		UpdatedAt: time.Now(),
	}
	if err := vinyl.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.service.UpdateVinyl(ctx, vinyl); err != nil {
		if errors.Is(err, domain.ErrVinylNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "vinyl not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, toVinylResponse(vinyl))
}

// DeleteVinyl godoc
// @Summary      Remover disco
// @Description  Remove um disco de vinil pelo seu ID
// @Tags         vinyls
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Vinyl ID"
// @Success      204
// @Failure      404 {object} ErrorResponse
// @Router       /vinyls/{id} [delete]
func (h *VinylHandler) DeleteVinyl(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.DeleteVinyl")
	defer span.End()

	id := c.Param("id")
	if err := h.service.DeleteVinyl(ctx, id); err != nil {
		if errors.Is(err, domain.ErrVinylNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "vinyl not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
```

**Step 4: Rodar testes**

```bash
go test ./internal/adapters/primary/http/... -v
```

Expected: PASS — 4 testes

**Step 5: Commit**

```bash
git add internal/adapters/primary/http/handler.go
git commit -m "feat: add HTTP handlers for vinyl CRUD and auth"
```

---

## Task 12: Adapter Primário — Router e Middlewares

**Files:**
- Create: `internal/adapters/primary/http/router.go`

**Step 1: Criar `internal/adapters/primary/http/router.go`**

```go
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

	// Auth
	auth := r.Group("/api/v1/auth")
	auth.POST("/login", handler.Login)

	// Vinyls (protegido por JWT)
	vinyls := r.Group("/api/v1/vinyls")
	vinyls.Use(jwtMiddleware(jwtSvc))
	{
		vinyls.GET("", handler.ListVinyls)
		vinyls.GET("/:id", handler.GetVinyl)
		vinyls.POST("", handler.CreateVinyl)
		vinyls.PUT("/:id", handler.UpdateVinyl)
		vinyls.DELETE("/:id", handler.DeleteVinyl)
	}

	return r
}

func jwtMiddleware(jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "authorization header required"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization format"})
			return
		}
		claims, err := jwtSvc.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
			return
		}
		c.Set("username", claims.Username)
		c.Next()
	}
}
```

**Step 2: Commit**

```bash
git add internal/adapters/primary/http/router.go
git commit -m "feat: add Gin router with JWT middleware, Prometheus and Swagger"
```

---

## Task 13: Entrypoint — main.go

**Files:**
- Create: `cmd/api/main.go`

**Step 1: Criar `cmd/api/main.go`**

```go
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
// @description Type "Bearer" followed by a space and JWT token.

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

	"github.com/nataliagranato/vinyl-catalog/internal/adapters/secondary/postgres"
	httpAdapter "github.com/nataliagranato/vinyl-catalog/internal/adapters/primary/http"
	"github.com/nataliagranato/vinyl-catalog/internal/application"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/auth"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/config"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/database"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/observability"

	_ "github.com/nataliagranato/vinyl-catalog/docs/swagger"
)

func main() {
	// Logger estruturado
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Config
	_ = godotenv.Load()
	cfg := config.Load()

	// Tracing
	ctx := context.Background()
	tp, err := observability.NewTracerProvider(ctx, cfg.OTELEndpoint, cfg.OTELServiceName)
	if err != nil {
		slog.Warn("tracing unavailable", "error", err)
	} else {
		defer func() { _ = tp.Shutdown(ctx) }()
	}

	// Database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Migrations
	if err := postgres.AutoMigrate(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Wiring de dependências
	repo := postgres.NewVinylRepository(db)
	svc := application.NewVinylService(repo)
	jwtSvc := auth.NewJWTServiceFromConfig(cfg.JWTSecret, cfg.JWTExpirationHours)
	handler := httpAdapter.NewVinylHandlerWithAuth(svc, jwtSvc, cfg.AdminUsername, cfg.AdminPassword)
	router := httpAdapter.NewRouter(handler, jwtSvc)

	// Servidor HTTP com graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		slog.Info("server starting", "port", cfg.AppPort)
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
```

**Step 2: Verificar que compila**

```bash
go build ./cmd/api/...
```

Expected: sem erros de compilação

**Step 3: Commit**

```bash
git add cmd/
git commit -m "feat: add main entrypoint with graceful shutdown and DI wiring"
```

---

## Task 14: Infraestrutura Docker

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `otel-collector-config.yaml`

**Step 1: Criar `Dockerfile`**

```dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vinyl-catalog ./cmd/api/main.go

# Stage 2: Run
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/vinyl-catalog .

EXPOSE 8080

USER nobody:nobody

ENTRYPOINT ["./vinyl-catalog"]
```

**Step 2: Criar `docker-compose.yml`**

```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      APP_PORT: "8080"
      APP_ENV: production
      DB_HOST: db
      DB_PORT: "5432"
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: vinyl_catalog
      DB_SSLMODE: disable
      JWT_SECRET: change-me-in-production
      JWT_EXPIRATION_HOURS: "24"
      ADMIN_USERNAME: admin
      ADMIN_PASSWORD: admin
      OTEL_EXPORTER_OTLP_ENDPOINT: otel-collector:4317
      OTEL_SERVICE_NAME: vinyl-catalog
    depends_on:
      db:
        condition: service_healthy
      otel-collector:
        condition: service_started
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: vinyl_catalog
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml:ro
    ports:
      - "4317:4317"   # OTLP gRPC
      - "4318:4318"   # OTLP HTTP
    depends_on:
      - jaeger
    restart: unless-stopped

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686" # UI
      - "14250:14250" # gRPC
    environment:
      COLLECTOR_OTLP_ENABLED: "true"
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    restart: unless-stopped

volumes:
  postgres_data:
  prometheus_data:
```

**Step 3: Criar `otel-collector-config.yaml`**

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024

exporters:
  otlp/jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

  prometheus:
    endpoint: "0.0.0.0:8889"

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/jaeger]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [prometheus]
```

**Step 4: Criar `prometheus.yml`**

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'vinyl-catalog'
    static_configs:
      - targets: ['app:8080']
    metrics_path: /metrics

  - job_name: 'otel-collector'
    static_configs:
      - targets: ['otel-collector:8889']
```

**Step 5: Commit**

```bash
git add Dockerfile docker-compose.yml otel-collector-config.yaml prometheus.yml
git commit -m "feat: add Docker multi-stage build and docker-compose with full observability stack"
```

---

## Task 15: Swagger — Gerar Documentação

**Step 1: Instalar swag CLI**

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

**Step 2: Gerar docs**

```bash
swag init -g cmd/api/main.go -o docs/swagger
```

Expected: arquivos `docs/swagger/docs.go`, `swagger.json`, `swagger.yaml` gerados

**Step 3: Verificar compilação com docs**

```bash
go build ./cmd/api/...
```

Expected: PASS

**Step 4: Commit**

```bash
git add docs/swagger/
git commit -m "docs: generate initial Swagger documentation"
```

---

## Task 16: README.md

**Files:**
- Create: `README.md`

**Step 1: Criar `README.md`**

```markdown
# Vinyl Catalog

Sistema de catalogação de discos de vinil com API REST em Go.

## Arquitetura

Clean Architecture + Hexagonal (Ports & Adapters):

```
adapters (HTTP, PostgreSQL) → application (use cases) → domain (entidades, interfaces)
```

## Stack

- **Go 1.22+** com Gin, GORM
- **PostgreSQL 16** como banco de dados
- **JWT** para autenticação
- **Swagger** para documentação da API
- **Prometheus + OpenTelemetry + Jaeger** para observabilidade
- **mise** para gestão de tasks
- **Docker + Docker Compose** para containerização

## Início Rápido

### Pré-requisitos

- [mise](https://mise.jdx.dev/) instalado
- Docker e Docker Compose

### Subir com Docker Compose

```bash
cp .env.example .env
mise run docker:up
```

A API estará disponível em:
- **API:** http://localhost:8080
- **Swagger:** http://localhost:8080/swagger/index.html
- **Prometheus:** http://localhost:9090
- **Jaeger UI:** http://localhost:16686

### Desenvolvimento local

```bash
cp .env.example .env
# edite .env com suas configurações locais

mise run docker:up   # sobe apenas db e infra
mise run dev         # sobe a aplicação
```

## Autenticação

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

Use o token retornado no header `Authorization: Bearer <token>`.

## Endpoints

| Método | Rota | Descrição | Auth |
|--------|------|-----------|------|
| POST | /api/v1/auth/login | Gera token JWT | Não |
| GET | /api/v1/health | Health check | Não |
| GET | /metrics | Métricas Prometheus | Não |
| GET | /api/v1/vinyls | Lista discos | Sim |
| GET | /api/v1/vinyls/:id | Busca disco | Sim |
| POST | /api/v1/vinyls | Cria disco | Sim |
| PUT | /api/v1/vinyls/:id | Atualiza disco | Sim |
| DELETE | /api/v1/vinyls/:id | Remove disco | Sim |

## Tasks mise

```bash
mise run dev          # hot reload
mise run build        # compila binário
mise run test         # testes
mise run swag         # gera Swagger
mise run docker:up    # sobe todos os serviços
mise run docker:down  # derruba todos os serviços
mise run lint         # linter
```

## Exemplo de Uso

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r .token)

# Criar disco
curl -X POST http://localhost:8080/api/v1/vinyls \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Kind of Blue","artist":"Miles Davis","year":1959,"genre":"Jazz","label":"Columbia"}'

# Listar discos
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/vinyls
```
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add README with architecture overview and usage guide"
```

---

## Task 17: Verificação Final

**Step 1: Rodar todos os testes**

```bash
go test ./... -v -cover
```

Expected: todos os testes PASS, cobertura > 60%

**Step 2: Verificar build**

```bash
go build ./...
```

Expected: sem erros

**Step 3: Verificar go vet**

```bash
go vet ./...
```

Expected: sem warnings

**Step 4: Subir stack completa**

```bash
mise run docker:up
```

Expected: todos os 5 serviços healthy

**Step 5: Smoke test da API**

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# Verificar Swagger
open http://localhost:8080/swagger/index.html
```

**Step 6: Commit final**

```bash
git add -A
git commit -m "chore: finalize vinyl catalog implementation"
```
