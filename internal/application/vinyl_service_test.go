package application_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/nataliagranato/vinyl-catalog/internal/application"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

type mockRepo struct {
	vinyls map[string]*domain.Vinyl
	err    error
}

func newMockRepo() *mockRepo {
	return &mockRepo{vinyls: make(map[string]*domain.Vinyl)}
}

func (m *mockRepo) Create(_ context.Context, v *domain.Vinyl) error {
	if m.err != nil {
		return m.err
	}
	m.vinyls[v.ID] = v
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id string) (*domain.Vinyl, error) {
	if m.err != nil {
		return nil, m.err
	}
	v, ok := m.vinyls[id]
	if !ok {
		return nil, domain.ErrVinylNotFound
	}
	return v, nil
}

func (m *mockRepo) FindAll(_ context.Context) ([]domain.Vinyl, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]domain.Vinyl, 0, len(m.vinyls))
	for _, v := range m.vinyls {
		result = append(result, *v)
	}
	return result, nil
}

func (m *mockRepo) Update(_ context.Context, v *domain.Vinyl) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.vinyls[v.ID]; !ok {
		return domain.ErrVinylNotFound
	}
	m.vinyls[v.ID] = v
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.vinyls[id]; !ok {
		return domain.ErrVinylNotFound
	}
	delete(m.vinyls, id)
	return nil
}

func TestVinylService_CreateAndGet(t *testing.T) {
	repo := newMockRepo()
	svc := application.NewVinylService(repo)
	ctx := context.Background()

	vinyl := &domain.Vinyl{
		ID:     "test-id",
		Title:  "Kind of Blue",
		Artist: "Miles Davis",
		Year:   1959,
		Genre:  "Jazz",
		Label:  "Columbia",
	}

	if err := svc.CreateVinyl(ctx, vinyl); err != nil {
		t.Fatalf("CreateVinyl() error = %v", err)
	}

	got, err := svc.GetVinyl(ctx, "test-id")
	if err != nil {
		t.Fatalf("GetVinyl() error = %v", err)
	}
	if got.Title != "Kind of Blue" {
		t.Errorf("expected title 'Kind of Blue', got %s", got.Title)
	}
}

func TestVinylService_GetNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := application.NewVinylService(repo)

	_, err := svc.GetVinyl(context.Background(), "not-found")
	if !errors.Is(err, domain.ErrVinylNotFound) {
		t.Errorf("expected ErrVinylNotFound, got %v", err)
	}
}

func TestVinylService_UpdateAndDelete(t *testing.T) {
	repo := newMockRepo()
	svc := application.NewVinylService(repo)
	ctx := context.Background()

	vinyl := &domain.Vinyl{ID: "id-1", Title: "Original", Artist: "Artist", Year: 2000, Genre: "Rock", Label: "Label"}
	_ = svc.CreateVinyl(ctx, vinyl)

	vinyl.Title = "Updated"
	if err := svc.UpdateVinyl(ctx, vinyl); err != nil {
		t.Fatalf("UpdateVinyl() error = %v", err)
	}

	if err := svc.DeleteVinyl(ctx, "id-1"); err != nil {
		t.Fatalf("DeleteVinyl() error = %v", err)
	}

	_, err := svc.GetVinyl(ctx, "id-1")
	if !errors.Is(err, domain.ErrVinylNotFound) {
		t.Errorf("expected ErrVinylNotFound after delete, got %v", err)
	}
}

func TestVinylService_ListVinyls(t *testing.T) {
	repo := newMockRepo()
	svc := application.NewVinylService(repo)
	ctx := context.Background()

	for i, title := range []string{"Album A", "Album B", "Album C"} {
		v := &domain.Vinyl{ID: fmt.Sprintf("id-%d", i), Title: title, Artist: "Artist", Year: 2000, Genre: "Pop", Label: "Label"}
		_ = svc.CreateVinyl(ctx, v)
	}

	vinyls, err := svc.ListVinyls(ctx)
	if err != nil {
		t.Fatalf("ListVinyls() error = %v", err)
	}
	if len(vinyls) != 3 {
		t.Errorf("expected 3 vinyls, got %d", len(vinyls))
	}
}
