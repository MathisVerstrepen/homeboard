package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"

	c "diikstra.fr/homeboard/cmd/cache"
	db "diikstra.fr/homeboard/cmd/database"
	mod "diikstra.fr/homeboard/cmd/home/modules"
	views "diikstra.fr/homeboard/views"
	f "github.com/MathisVerstrepen/go-module/webfetch"

	. "diikstra.fr/homeboard/cmd/models"
)

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbConn := db.Connect()
	cache := c.Connect()

	homeLayout := HomeLayoutData{
		NRows:  3,
		NCols:  3,
		Layout: dbConn.GetHomeLayouts(),
	}
	fmt.Println(*homeLayout.Layout)

	proxies := f.InitFetchers(basepath)
	moduleService := mod.ModuleService{
		Proxies: proxies,
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	e.Static("/images", "images")
	e.Static("/css", "css")

	background, err := dbConn.GetSelectedBackground()
	if err != nil {
		log.Printf("%v", err)
		log.Fatal("Fail to fetch selected background")
	}
	globalPageData := PageData{
		Background: background,
	}

	e.GET("/", func(c echo.Context) error {
		newPageData := PageData{
			Title:      "Home",
			Page:       "home",
			Background: globalPageData.Background,
			HomeLayout: homeLayout,
		}

		return Render(c, http.StatusOK, views.Root(views.Home(homeLayout), newPageData))
	})

	e.GET("/home/modules", func(c echo.Context) error {
		modules := moduleService.GetModulesMetadata()
		for _, modulePosition := range *homeLayout.Layout {
			for _, module := range modules {
				if modulePosition.ModuleName == module.Name {
					statusCode, component, error := moduleService.RenderModule(cache, module.Name, modulePosition.Position)

					if error != nil {
						Render(c, http.StatusBadRequest, nil)
					} else {
						Render(c, statusCode, component)
					}
				}
			}
		}

		return nil
	})

	e.POST("/home/modules/:moduleName/:position", func(c echo.Context) error {
		moduleName := c.Param("moduleName")
		position := c.Param("position")

		err := dbConn.SetHomeLayout(position, moduleName)
		if err != nil {
			return err
		}

		homeLayout.Layout = dbConn.GetHomeLayouts()

		statusCode, component, error := moduleService.RenderModule(cache, moduleName, position)
		if error != nil {
			return Render(c, http.StatusBadRequest, nil)
		} else {
			return Render(c, statusCode, component)
		}
	})

	e.GET("/home/edit", func(c echo.Context) error {
		Render(c, http.StatusOK, views.Header_buttons_out())

		nCols := homeLayout.NCols
		nRows := homeLayout.NRows
		ids := make([]string, 0)
		for row := range nRows {
			for col := range nCols {
				position := fmt.Sprintf("card_%d_%d", row+1, col+1)

				present := false
				for _, layout := range *homeLayout.Layout {
					if layout.Position == position {
						present = true
						break
					}
				}

				if !present {
					ids = append(ids, position)
				}
			}
		}

		return Render(c, http.StatusOK, views.BlockEdit(ids))
	})

	e.POST("/home/edit", func(c echo.Context) error {
		Render(c, http.StatusOK, views.Header_buttons())
		return Render(c, http.StatusOK, views.HomeLayout(homeLayout))
	})

	e.GET("/home/add/list/:position", func(c echo.Context) error {
		addPopupData := HomeAddPopup{
			Position: c.Param("position"),
			Modules:  moduleService.GetModulesMetadata(),
		}

		return Render(c, http.StatusOK, views.AddBlockPopup(addPopupData))
	})

	e.GET("/settings", func(c echo.Context) error {
		newPageData := PageData{
			Title:      "Settings",
			Page:       "settings",
			Background: globalPageData.Background,
		}

		return Render(c, http.StatusOK, views.Root(views.Settings(), newPageData))
	})

	e.GET("/settings/backgrounds", func(c echo.Context) error {
		return Render(c, http.StatusOK, views.BgPopup(BackgroundData{
			Backgrounds: dbConn.GetBackgrounds(),
		}))
	})

	e.POST("/settings/backgrounds", func(c echo.Context) error {
		bg, err := dbConn.UploadBackground(c)
		if err != nil {
			return err
		}

		return Render(c, http.StatusOK, views.BgItem(bg))
	})

	e.POST("/settings/backgrounds/selected/:id", func(c echo.Context) error {
		idBg := c.Param("id")
		id, err := strconv.Atoi(idBg)
		if err != nil {
			return c.String(400, "Invalid id")
		}

		Render(c, http.StatusOK, views.OobButtonBgSelect(globalPageData.Background))

		background, err = dbConn.SetSelectedBackground(id)
		if err != nil {
			return c.String(400, "Fail to set new background :"+err.Error())
		}
		globalPageData.Background = background

		Render(c, http.StatusOK, views.OobButtonBgSelected(background))
		return Render(c, http.StatusOK, views.Background(globalPageData))
	})

	e.DELETE("/settings/backgrounds/:id", func(c echo.Context) error {
		idBg := c.Param("id")
		id, err := strconv.Atoi(idBg)
		if err != nil {
			return nil
		}

		err = dbConn.DeleteBackground(id)
		if err != nil {
			return nil
		}

		return nil
	})

	e.GET("/ping", func(c echo.Context) error {
		return c.String(200, "pong")
	})

	e.GET("/ws", func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()
			for {
				msg := ""
				err := websocket.Message.Receive(ws, &msg)
				if err != nil {
					return
				}
			}
		}).ServeHTTP(c.Response(), c.Request())
		return nil
	})

	e.Logger.Fatal(e.Start(":42069"))
}
