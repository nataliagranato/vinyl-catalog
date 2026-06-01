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
	if cfg.JWTExpirationHours != "24" {
		t.Errorf("expected default expiration 24, got %s", cfg.JWTExpirationHours)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_HOST", "myhost")
	os.Setenv("JWT_SECRET", "supersecret")
	defer os.Clearenv()

	cfg := config.Load()

	if cfg.AppPort != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.AppPort)
	}
	if cfg.DBHost != "myhost" {
		t.Errorf("expected host myhost, got %s", cfg.DBHost)
	}
	if cfg.JWTSecret != "supersecret" {
		t.Errorf("expected jwt secret supersecret, got %s", cfg.JWTSecret)
	}
}
