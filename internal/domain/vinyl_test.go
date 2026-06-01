package domain_test

import (
	"testing"
	"time"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

func TestVinyl_Validate(t *testing.T) {
	tests := []struct {
		name    string
		vinyl   domain.Vinyl
		wantErr bool
	}{
		{
			name: "valid vinyl",
			vinyl: domain.Vinyl{
				Title:  "Kind of Blue",
				Artist: "Miles Davis",
				Year:   1959,
				Genre:  "Jazz",
				Label:  "Columbia",
			},
			wantErr: false,
		},
		{
			name:    "missing title",
			vinyl:   domain.Vinyl{Artist: "Miles Davis", Year: 1959},
			wantErr: true,
		},
		{
			name:    "missing artist",
			vinyl:   domain.Vinyl{Title: "Kind of Blue", Year: 1959},
			wantErr: true,
		},
		{
			name:    "year too old",
			vinyl:   domain.Vinyl{Title: "T", Artist: "A", Year: 1800},
			wantErr: true,
		},
		{
			name:    "year boundary lower valid",
			vinyl:   domain.Vinyl{Title: "T", Artist: "A", Year: 1860},
			wantErr: false,
		},
		{
			name:    "year boundary upper valid",
			vinyl:   domain.Vinyl{Title: "T", Artist: "A", Year: time.Now().Year() + 1},
			wantErr: false,
		},
		{
			name:    "future year invalid",
			vinyl:   domain.Vinyl{Title: "T", Artist: "A", Year: time.Now().Year() + 2},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vinyl.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewVinyl(t *testing.T) {
	v, err := domain.NewVinyl("test-id", "Kind of Blue", "Miles Davis", 1959, "Jazz", "Columbia")
	if err != nil {
		t.Fatalf("NewVinyl() unexpected error: %v", err)
	}
	if v.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %s", v.ID)
	}
	if v.Title != "Kind of Blue" {
		t.Errorf("expected title 'Kind of Blue', got %s", v.Title)
	}
	if v.CreatedAt.IsZero() {
		t.Error("NewVinyl() should set CreatedAt")
	}
	if v.UpdatedAt.IsZero() {
		t.Error("NewVinyl() should set UpdatedAt")
	}
}

func TestNewVinyl_Invalid(t *testing.T) {
	_, err := domain.NewVinyl("id", "", "Miles Davis", 1959, "Jazz", "Columbia")
	if err == nil {
		t.Error("NewVinyl() with empty title should return error")
	}
}
