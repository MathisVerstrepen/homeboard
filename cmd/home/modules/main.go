package modules

import "github.com/labstack/echo/v4"

type ModuleMetada struct {
	Name  string
	Icon  string
	Sizes []string
}

type HomeModule interface {
	GetMetadata() ModuleMetada
	RenderView(echo.Context, string)
}
