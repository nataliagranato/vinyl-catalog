package domain

import "context"

type Profile struct {
	ID               string   `json:"id"`
	Username         string   `json:"username"`
	DisplayName      string   `json:"display_name"`
	Bio              string   `json:"bio"`
	PhotoURL         string   `json:"photo_url"`
	Links            []string `json:"links"`
	PreferredGenres  []string `json:"preferred_genres"`
	FavoriteVinylIDs []string `json:"favorite_vinyl_ids"`
}

type ProfileRepository interface {
	Get(ctx context.Context) (*Profile, error)
	Upsert(ctx context.Context, p *Profile) error
}
