package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nataliagranato/vinyl-catalog/internal/application"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

type mockTrackRepo struct {
	tracks map[string]*domain.Track
	err    error
}

func newMockTrackRepo() *mockTrackRepo {
	return &mockTrackRepo{tracks: make(map[string]*domain.Track)}
}

func (m *mockTrackRepo) Create(_ context.Context, t *domain.Track) error {
	if m.err != nil {
		return m.err
	}
	m.tracks[t.ID] = t
	return nil
}

func (m *mockTrackRepo) FindByVinylID(_ context.Context, vinylID string) ([]domain.Track, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Track
	for _, t := range m.tracks {
		if t.VinylID == vinylID {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (m *mockTrackRepo) FindByID(_ context.Context, id string) (*domain.Track, error) {
	if m.err != nil {
		return nil, m.err
	}
	t, ok := m.tracks[id]
	if !ok {
		return nil, domain.ErrTrackNotFound
	}
	return t, nil
}

func (m *mockTrackRepo) Update(_ context.Context, t *domain.Track) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.tracks[t.ID]; !ok {
		return domain.ErrTrackNotFound
	}
	m.tracks[t.ID] = t
	return nil
}

func (m *mockTrackRepo) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.tracks[id]; !ok {
		return domain.ErrTrackNotFound
	}
	delete(m.tracks, id)
	return nil
}

func TestTrackService_CreateAndList(t *testing.T) {
	repo := newMockTrackRepo()
	svc := application.NewTrackService(repo)
	ctx := context.Background()

	track := &domain.Track{ID: "t1", VinylID: "v1", Title: "So What", Position: 1}
	if err := svc.CreateTrack(ctx, track); err != nil {
		t.Fatalf("CreateTrack() error = %v", err)
	}

	tracks, err := svc.ListTracks(ctx, "v1")
	if err != nil {
		t.Fatalf("ListTracks() error = %v", err)
	}
	if len(tracks) != 1 {
		t.Errorf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Title != "So What" {
		t.Errorf("expected title 'So What', got %s", tracks[0].Title)
	}
}

func TestTrackService_UpdateAndDelete(t *testing.T) {
	repo := newMockTrackRepo()
	svc := application.NewTrackService(repo)
	ctx := context.Background()

	track := &domain.Track{ID: "t1", VinylID: "v1", Title: "Original", Position: 1}
	_ = svc.CreateTrack(ctx, track)

	track.Title = "Updated"
	if err := svc.UpdateTrack(ctx, track); err != nil {
		t.Fatalf("UpdateTrack() error = %v", err)
	}

	if err := svc.DeleteTrack(ctx, "t1"); err != nil {
		t.Fatalf("DeleteTrack() error = %v", err)
	}
}

func TestTrackService_DeleteNotFound(t *testing.T) {
	repo := newMockTrackRepo()
	svc := application.NewTrackService(repo)

	err := svc.DeleteTrack(context.Background(), "nonexistent")
	if !errors.Is(err, domain.ErrTrackNotFound) {
		t.Errorf("expected ErrTrackNotFound, got %v", err)
	}
}

func TestTrackService_ListEmpty(t *testing.T) {
	repo := newMockTrackRepo()
	svc := application.NewTrackService(repo)

	tracks, err := svc.ListTracks(context.Background(), "no-such-vinyl")
	if err != nil {
		t.Fatalf("ListTracks() error = %v", err)
	}
	if len(tracks) != 0 {
		t.Errorf("expected empty list, got %d", len(tracks))
	}
}
