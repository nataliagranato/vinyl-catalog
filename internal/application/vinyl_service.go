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
