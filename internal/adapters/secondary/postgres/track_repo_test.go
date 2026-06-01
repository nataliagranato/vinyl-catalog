package postgres_test

import (
	"context"
	"testing"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nataliagranato/vinyl-catalog/internal/adapters/secondary/postgres"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=vinyl_catalog_test sslmode=disable"
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("PostgreSQL not available, skipping integration test: %v", err)
	}
	if err := postgres.AutoMigrateAll(db); err != nil {
		t.Fatalf("AutoMigrateAll error: %v", err)
	}
	return db
}

func TestTrackRepository_Implements(t *testing.T) {
	var _ domain.TrackRepository = (*postgres.TrackRepository)(nil)
	t.Log("TrackRepository correctly implements domain.TrackRepository interface")
}

func TestTrackRepository_CRUD(t *testing.T) {
	db := openTestDB(t)
	repo := postgres.NewTrackRepository(db)
	ctx := context.Background()

	// Criar vinyl primeiro (FK constraint)
	vinylRepo := postgres.NewVinylRepository(db)
	vinyl := &domain.Vinyl{ID: "test-vinyl-track-crud", Title: "Test", Artist: "Artist", Year: 2000, Genre: "Rock", Label: "Label"}
	if err := vinylRepo.Create(ctx, vinyl); err != nil {
		t.Fatalf("create vinyl: %v", err)
	}
	t.Cleanup(func() { vinylRepo.Delete(ctx, vinyl.ID) })

	// Create
	track := &domain.Track{ID: "track-crud-test", VinylID: vinyl.ID, Title: "Track 1", Position: 1}
	if err := repo.Create(ctx, track); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Cleanup(func() { repo.Delete(ctx, track.ID) })

	// FindByVinylID
	tracks, err := repo.FindByVinylID(ctx, vinyl.ID)
	if err != nil {
		t.Fatalf("FindByVinylID() error = %v", err)
	}
	if len(tracks) == 0 {
		t.Fatal("expected at least 1 track")
	}

	// FindByID
	got, err := repo.FindByID(ctx, track.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Title != "Track 1" {
		t.Errorf("expected 'Track 1', got %s", got.Title)
	}

	// Update
	track.Title = "Updated Track"
	if err := repo.Update(ctx, track); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Delete
	if err := repo.Delete(ctx, track.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify gone
	_, err = repo.FindByID(ctx, track.ID)
	if err != domain.ErrTrackNotFound {
		t.Errorf("expected ErrTrackNotFound after delete, got %v", err)
	}
}
