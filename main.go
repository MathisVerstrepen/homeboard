package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"diikstra.fr/homeboard/handlers"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	e.Static("/assets", "assets")

	handlers.Init()

	// ---- Home Routes ---- //
	e.GET("/", handlers.HomeHandler)
	e.GET("/home/edit", handlers.HomeGetEdit)
	e.POST("/home/edit", handlers.HomePostEdit)
	e.GET("/home/add/list/:position", handlers.HomeGetAddList)
	e.GET("/home/modules", handlers.HomeModulesHandler)
	e.GET("/home/module/:moduleName/:position", handlers.HomeModuleHandler)
	e.DELETE("/home/module/:moduleName/:position", handlers.HomeModuleDelete)
	e.POST("/home/modules/:moduleName/:position", handlers.HomeAddModulePositionHandler)
	e.GET("/home/modules/edit/:moduleName/:position", handlers.HomeGetModuleEdit)

	// ---- Settings Routes ---- //
	e.GET("/settings", handlers.SettingsHandler)
	e.GET("/settings/backgrounds", handlers.SettingsGetBackgrounds)
	e.POST("/settings/backgrounds", handlers.SettingsPostBackground)
	e.POST("/settings/backgrounds/selected/:id", handlers.SettingsSetSelectedBackgroundfunc)
	e.DELETE("/settings/backgrounds/:id", handlers.SettingsDeleteBackground)

	// ---- Global Routes ---- //
	e.GET("/ping", handlers.GlobalPing)
	e.GET("/ws", handlers.GlobalHotReloadWS)

	e.Logger.Fatal(e.Start(":42069"))
}
