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

	// Add Module PopUp
	e.GET("/home/edit", handlers.HomeGetEdit)
	e.POST("/home/edit", handlers.HomePostEdit)
	e.GET("/home/add/list/:position", handlers.HomeGetAddList)

	// Home Layout
	e.GET("/home/modules", handlers.HomeModulesHandler)
	e.GET("/home/module/:moduleName/:position", handlers.HomeModuleHandler)
	e.GET("/home/module/refresh/:moduleName/:position", handlers.HomeModuleHandlerForceRefresh)
	e.POST("/home/modules/:moduleName/:position", handlers.HomeAddModulePositionHandler)
	e.DELETE("/home/module/:moduleName/:position", handlers.HomeModuleDelete)

	// Module Settings
	e.GET("/home/module/edit/:moduleName/:position", handlers.HomeGetModuleEdit)
	e.GET("/home/module/edit/:moduleName/:position/variables", handlers.HomeGetModuleEditVariables)
	e.POST("/home/module/edit/:moduleName/:position/variables", handlers.HomePostModuleEditVariables)

	// ---- Settings Routes ---- //
	e.GET("/settings", handlers.SettingsHandler)

	// Background
	e.GET("/settings/backgrounds", handlers.SettingsGetBackgrounds)
	e.POST("/settings/backgrounds", handlers.SettingsPostBackground)
	e.POST("/settings/backgrounds/selected/:id", handlers.SettingsSetSelectedBackgroundfunc)
	e.DELETE("/settings/backgrounds/:id", handlers.SettingsDeleteBackground)

	// ---- Linkhub Routes ---- //
	e.GET("/linkhub", handlers.LinkHub)
	e.GET("/linkhub/:id", handlers.LinkHubId)
	e.GET("/linkhub/image/:id", handlers.LinkHubImageDetails)
	e.POST("/linkhub/site", handlers.LinkHubPostSite)

	// ---- Global Routes ---- //
	e.GET("/ping", handlers.GlobalPing)
	e.GET("/ws", handlers.GlobalHotReloadWS)

	e.Logger.Fatal(e.Start(":42069"))
}
