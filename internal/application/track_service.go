package application

import (
	"context"

	"go.opentelemetry.io/otel"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

type trackService struct {
	repo domain.TrackRepository
}

func NewTrackService(repo domain.TrackRepository) domain.TrackService {
	return &trackService{repo: repo}
}

func (s *trackService) CreateTrack(ctx context.Context, t *domain.Track) error {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackService.CreateTrack")
	defer span.End()
	return s.repo.Create(ctx, t)
}

func (s *trackService) ListTracks(ctx context.Context, vinylID string) ([]domain.Track, error) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackService.ListTracks")
	defer span.End()
	return s.repo.FindByVinylID(ctx, vinylID)
}

func (s *trackService) UpdateTrack(ctx context.Context, t *domain.Track) error {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackService.UpdateTrack")
	defer span.End()
	return s.repo.Update(ctx, t)
}

func (s *trackService) DeleteTrack(ctx context.Context, id string) error {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackService.DeleteTrack")
	defer span.End()
	return s.repo.Delete(ctx, id)
}
