package postgres_test

import (
	"context"
	"testing"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nataliagranato/vinyl-catalog/internal/adapters/secondary/postgres"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

// Compile-time check: VinylRepository implements domain.VinylRepository
var _ domain.VinylRepository = (*postgres.VinylRepository)(nil)

func TestVinylRepository_ImplementsInterface(t *testing.T) {
	t.Log("VinylRepository correctly implements domain.VinylRepository interface")
}

func TestVinylRepository_NotFound(t *testing.T) {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=vinyl_catalog_test sslmode=disable"
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("PostgreSQL not available, skipping integration test: %v", err)
	}

	if err := postgres.AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate error: %v", err)
	}

	repo := postgres.NewVinylRepository(db)
	_, err = repo.FindByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != domain.ErrVinylNotFound {
		t.Errorf("expected ErrVinylNotFound, got %v", err)
	}
}
