package models

import (
	db "diikstra.fr/homeboard/db/database"
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
