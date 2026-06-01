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
	Genre       string
	Label       string
	Description string
	CoverURL    string `gorm:"column:cover_url"`
	CreatedAt   time.Time
	UpdatedAt time.Time
}

func (vinylModel) TableName() string { return "vinyls" }

func toModel(v *domain.Vinyl) *vinylModel {
	return &vinylModel{
		ID:        v.ID,
		Title:     v.Title,
		Artist:    v.Artist,
		Year:      v.Year,
		Genre:       v.Genre,
		Label:       v.Label,
		Description: v.Description,
		CoverURL:    v.CoverURL,
		CreatedAt:   v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

func toDomain(m *vinylModel) *domain.Vinyl {
	return &domain.Vinyl{
		ID:        m.ID,
		Title:     m.Title,
		Artist:    m.Artist,
		Year:      m.Year,
		Genre:       m.Genre,
		Label:       m.Label,
		Description: m.Description,
		CoverURL:    m.CoverURL,
		CreatedAt:   m.CreatedAt,
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
		vinyls[i] = *toDomain(&m)
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
