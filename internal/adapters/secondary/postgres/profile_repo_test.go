package postgres_test

import (
	"context"
	"testing"

	"github.com/nataliagranato/vinyl-catalog/internal/adapters/secondary/postgres"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

func TestProfileRepository_Implements(t *testing.T) {
	var _ domain.ProfileRepository = (*postgres.ProfileRepository)(nil)
	t.Log("ProfileRepository correctly implements domain.ProfileRepository interface")
}

func TestProfileRepository_GetEmpty(t *testing.T) {
	db := openTestDB(t)
	repo := postgres.NewProfileRepository(db)

	profile, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
}

func TestProfileRepository_Upsert(t *testing.T) {
	db := openTestDB(t)
	repo := postgres.NewProfileRepository(db)
	ctx := context.Background()

	p := &domain.Profile{
		Username:        "admin",
		DisplayName:     "Natalia",
		Bio:             "Colecionadora de jazz.",
		Links:           []string{"https://github.com/nataliagranato"},
		PreferredGenres: []string{"Jazz", "Soul"},
	}

	if err := repo.Upsert(ctx, p); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("Get() after Upsert error = %v", err)
	}
	if got.DisplayName != "Natalia" {
		t.Errorf("expected DisplayName='Natalia', got %s", got.DisplayName)
	}
	if len(got.Links) == 0 {
		t.Error("expected links to be populated")
	}
}
