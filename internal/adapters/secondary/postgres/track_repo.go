package postgres

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"gorm.io/gorm"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

type trackModel struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	VinylID   string    `gorm:"not null;index"`
	Title     string    `gorm:"not null"`
	Position  int
	Lyrics    string    `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (trackModel) TableName() string { return "tracks" }

func toTrackDomain(m *trackModel) *domain.Track {
	return &domain.Track{
		ID:        m.ID,
		VinylID:   m.VinylID,
		Title:     m.Title,
		Position:  m.Position,
		Lyrics:    m.Lyrics,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type TrackRepository struct {
	db *gorm.DB
}

func NewTrackRepository(db *gorm.DB) *TrackRepository {
	return &TrackRepository{db: db}
}

// AutoMigrateAll runs migrations for all models.
// profileModel will be added in a later task.
func AutoMigrateAll(db *gorm.DB) error {
	return db.AutoMigrate(&vinylModel{}, &trackModel{}, &profileModel{})
}

func (r *TrackRepository) Create(ctx context.Context, t *domain.Track) error {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackRepository.Create")
	defer span.End()
	m := &trackModel{
		ID: t.ID, VinylID: t.VinylID, Title: t.Title,
		Position: t.Position, Lyrics: t.Lyrics,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *TrackRepository) FindByVinylID(ctx context.Context, vinylID string) ([]domain.Track, error) {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackRepository.FindByVinylID")
	defer span.End()
	var models []trackModel
	if err := r.db.WithContext(ctx).Where("vinyl_id = ?", vinylID).Order("position asc").Find(&models).Error; err != nil {
		return nil, err
	}
	tracks := make([]domain.Track, len(models))
	for i, m := range models {
		tracks[i] = *toTrackDomain(&m)
	}
	return tracks, nil
}

func (r *TrackRepository) FindByID(ctx context.Context, id string) (*domain.Track, error) {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackRepository.FindByID")
	defer span.End()
	var m trackModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTrackNotFound
	}
	if err != nil {
		return nil, err
	}
	return toTrackDomain(&m), nil
}

func (r *TrackRepository) Update(ctx context.Context, t *domain.Track) error {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackRepository.Update")
	defer span.End()
	result := r.db.WithContext(ctx).Model(&trackModel{}).Where("id = ?", t.ID).Updates(map[string]any{
		"title": t.Title, "position": t.Position, "lyrics": t.Lyrics, "updated_at": t.UpdatedAt,
	})
	if result.RowsAffected == 0 {
		return domain.ErrTrackNotFound
	}
	return result.Error
}

func (r *TrackRepository) Delete(ctx context.Context, id string) error {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackRepository.Delete")
	defer span.End()
	result := r.db.WithContext(ctx).Delete(&trackModel{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return domain.ErrTrackNotFound
	}
	return result.Error
}
