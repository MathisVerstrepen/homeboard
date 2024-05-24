package models

import (
	db "diikstra.fr/homeboard/cmd/database"
)

type PageData struct {
	Title      string
	Page       string
	Background db.Background
	HomeLayout HomeLayoutData
}

type BackgroundData struct {
	Backgrounds *[]db.Background
}

type HomeLayoutData struct {
	NRows  int
	NCols  int
	Layout *[]db.ModulePosition
}

type ModuleMetada struct {
	Name     string
	Icon     string
	Sizes    []string
	Position string
	CacheKey string
}

type HomeAddPopup struct {
	Position string
	Modules  []ModuleMetada
}

type RenderData struct {
	MovieData []MovieData
	Metadata  ModuleMetada
}

type MovieData struct {
	Poster      string
	Owner       string
	OwnerAvatar string
	Rating      string
	Slug        string
	Id          string
}
