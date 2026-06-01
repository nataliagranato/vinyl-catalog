package domain

import (
	"fmt"
	"time"
)

type Vinyl struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Artist    string    `json:"artist"`
	Year      int       `json:"year"`
	Genre     string    `json:"genre"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	CoverURL    string    `json:"cover_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewVinyl(id, title, artist string, year int, genre, label string) (*Vinyl, error) {
	now := time.Now()
	v := &Vinyl{
		ID:        id,
		Title:     title,
		Artist:    artist,
		Year:      year,
		Genre:     genre,
		Label:     label,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *Vinyl) Validate() error {
	if v.Title == "" {
		return ErrVinylTitleEmpty
	}
	if v.Artist == "" {
		return ErrVinylArtistEmpty
	}
	currentYear := time.Now().Year()
	if v.Year < 1860 || v.Year > currentYear+1 {
		return fmt.Errorf("%w: got %d", ErrVinylYearInvalid, v.Year)
	}
	return nil
}
