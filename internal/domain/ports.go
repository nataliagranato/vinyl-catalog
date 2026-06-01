package domain

import "context"

type VinylRepository interface {
	Create(ctx context.Context, v *Vinyl) error
	FindByID(ctx context.Context, id string) (*Vinyl, error)
	FindAll(ctx context.Context) ([]Vinyl, error)
	Update(ctx context.Context, v *Vinyl) error
	Delete(ctx context.Context, id string) error
}

type VinylService interface {
	CreateVinyl(ctx context.Context, v *Vinyl) error
	GetVinyl(ctx context.Context, id string) (*Vinyl, error)
	ListVinyls(ctx context.Context) ([]Vinyl, error)
	UpdateVinyl(ctx context.Context, v *Vinyl) error
	DeleteVinyl(ctx context.Context, id string) error
}
