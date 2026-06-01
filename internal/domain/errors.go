package domain

import "errors"

var (
	ErrVinylNotFound    = errors.New("vinyl not found")
	ErrVinylTitleEmpty  = errors.New("vinyl title is required")
	ErrVinylArtistEmpty = errors.New("vinyl artist is required")
	ErrVinylYearInvalid = errors.New("vinyl year must be between 1860 and current year")
	ErrTrackNotFound    = errors.New("track not found")
)
