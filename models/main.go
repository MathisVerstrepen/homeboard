package models

import (
	db "diikstra.fr/homeboard/pkg/db"
	"github.com/a-h/templ"
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

type HomeAddPopup struct {
	Position string
	Modules  []ModuleMetada
}

type ButtonMeta struct {
	Icon       templ.Component
	Target     string
	Id         string
	Swap       bool
	PostAction string
	Include    string
}
