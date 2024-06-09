package models

type LetterboxdMovieData struct {
	Poster      string
	Owner       string
	OwnerAvatar string
	Rating      string
	Slug        string
	Id          string
}

type LetterboxdRenderData struct {
	MovieData []LetterboxdMovieData
	Metadata  ModuleMetada
	Data      ModuleData
}
