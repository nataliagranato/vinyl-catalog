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
	svc := auth.NewJWTService("test-secret", -1)
	token, _ := svc.GenerateToken("admin")
	time.Sleep(10 * time.Millisecond)
	_, err := svc.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestNewJWTServiceFromConfig(t *testing.T) {
	svc := auth.NewJWTServiceFromConfig("my-secret", "2")
	token, err := svc.GenerateToken("user")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.Username != "user" {
		t.Errorf("expected username user, got %s", claims.Username)
	}
}
