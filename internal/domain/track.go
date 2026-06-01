package domain

import (
	"context"
	"errors"
	"time"
)

type Track struct {
	ID        string    `json:"id"`
	VinylID   string    `json:"vinyl_id"`
	Title     string    `json:"title"`
	Position  int       `json:"position"`
	Lyrics    string    `json:"lyrics"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *Track) Validate() error {
	if t.Title == "" {
		return errors.New("track title is required")
	}
	if t.VinylID == "" {
		return errors.New("track vinyl_id is required")
	}
	return nil
}

type TrackRepository interface {
	Create(ctx context.Context, t *Track) error
	FindByVinylID(ctx context.Context, vinylID string) ([]Track, error)
	FindByID(ctx context.Context, id string) (*Track, error)
	Update(ctx context.Context, t *Track) error
	Delete(ctx context.Context, id string) error
}

type TrackService interface {
	CreateTrack(ctx context.Context, t *Track) error
	ListTracks(ctx context.Context, vinylID string) ([]Track, error)
	UpdateTrack(ctx context.Context, t *Track) error
	DeleteTrack(ctx context.Context, id string) error
}
