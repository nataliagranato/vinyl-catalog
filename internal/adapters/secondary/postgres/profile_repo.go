package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

type profileModel struct {
	ID               string `gorm:"primaryKey;type:uuid"`
	Username         string `gorm:"uniqueIndex;not null"`
	DisplayName      string
	Bio              string `gorm:"type:text"`
	PhotoURL         string `gorm:"column:photo_url"`
	Links            string `gorm:"type:text"` // comma-separated
	PreferredGenres  string `gorm:"type:text"` // comma-separated
	FavoriteVinylIDs string `gorm:"type:text;column:favorite_vinyl_ids"` // comma-separated
}

func (profileModel) TableName() string { return "profiles" }

func splitCSV(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}

func joinCSV(ss []string) string {
	return strings.Join(ss, ",")
}

func toProfileDomain(m *profileModel) *domain.Profile {
	return &domain.Profile{
		ID:               m.ID,
		Username:         m.Username,
		DisplayName:      m.DisplayName,
		Bio:              m.Bio,
		PhotoURL:         m.PhotoURL,
		Links:            splitCSV(m.Links),
		PreferredGenres:  splitCSV(m.PreferredGenres),
		FavoriteVinylIDs: splitCSV(m.FavoriteVinylIDs),
	}
}

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Get(ctx context.Context) (*domain.Profile, error) {
	var m profileModel
	err := r.db.WithContext(ctx).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Return empty profile if none exists yet
		return &domain.Profile{Username: "admin", Links: []string{}, PreferredGenres: []string{}}, nil
	}
	if err != nil {
		return nil, err
	}
	return toProfileDomain(&m), nil
}

func (r *ProfileRepository) Upsert(ctx context.Context, p *domain.Profile) error {
	m := &profileModel{
		ID:               p.ID,
		Username:         p.Username,
		DisplayName:      p.DisplayName,
		Bio:              p.Bio,
		PhotoURL:         p.PhotoURL,
		Links:            joinCSV(p.Links),
		PreferredGenres:  joinCSV(p.PreferredGenres),
		FavoriteVinylIDs: joinCSV(p.FavoriteVinylIDs),
	}
	return r.db.WithContext(ctx).Save(m).Error
}
