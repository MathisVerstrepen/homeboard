package main

import (
	"html/template"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"

	db "diikstra.fr/homeboard/cmd/database"
	mod "diikstra.fr/homeboard/cmd/home/modules"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl := template.Must(t.templates.Clone())

	// Case when we need to render a block define in a html file
	// name syntax is <filename.html>/<blockname>
	if strings.Contains(name, "/") {
		names := strings.Split(name, "/")
		tmpl = template.Must(tmpl.ParseGlob("views/" + names[0]))
		return tmpl.ExecuteTemplate(w, names[1], data)
	}

	tmpl = template.Must(tmpl.ParseGlob("views/" + name))
	return tmpl.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	// Define fonctions available in html template context
	funcMap := template.FuncMap{
		"eq": func(a, b interface{}) bool {
			return a == b
		},
	}

	// https://stackoverflow.com/questions/36617949/how-to-use-base-template-file-for-golang-html-template
	templates := template.Must(template.New("").Funcs(funcMap).ParseGlob("views/*.html"))
	return &Templates{
		templates: templates,
	}
}

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
	Blocks []mod.HomeModule
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	conn := db.Connect()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Static("/images", "images")
	e.Static("/css", "css")

	e.Renderer = newTemplate()

	background, err := conn.GetSelectedBackground()
	if err != nil {
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
			HomeLayout: HomeLayoutData{
				NRows:  5,
				NCols:  3,
				Blocks: []mod.HomeModule{},
			},
		}
		return c.Render(200, "home.html", &newPageData)
	})

	e.GET("/settings", func(c echo.Context) error {
		newPageData := PageData{
			Title:      "Settings",
			Page:       "settings",
			Background: globalPageData.Background,
		}
		return c.Render(200, "settings.html", &newPageData)
	})

	e.GET("/settings/backgrounds", func(c echo.Context) error {
		return c.Render(200, "settings.html/bg-popup", BackgroundData{
			Backgrounds: conn.GetBackgrounds(),
		})
	})

	e.POST("/settings/backgrounds", func(c echo.Context) error {
		bg, err := conn.UploadBackground(c)
		if err != nil {
			return err
		}

		return c.Render(200, "settings.html/bg-item", bg)
	})

	e.POST("/settings/backgrounds/selected/:id", func(c echo.Context) error {
		idBg := c.Param("id")
		id, err := strconv.Atoi(idBg)
		if err != nil {
			return c.String(400, "Invalid id")
		}

		c.Render(200, "settings.html/oob-button-bg-select", globalPageData.Background)

		background, err = conn.SetSelectedBackground(id)
		if err != nil {
			return c.String(400, "Fail to set new background :"+err.Error())
		}
		globalPageData.Background = background

		c.Render(200, "settings.html/oob-button-bg-selected", background)
		return c.Render(200, "layout.html/background", globalPageData)
	})

	e.DELETE("/settings/backgrounds/:id", func(c echo.Context) error {
		idBg := c.Param("id")
		id, err := strconv.Atoi(idBg)
		if err != nil {
			return c.String(400, "Invalid id")
		}

		err = conn.DeleteBackground(id)
		if err != nil {
			return c.String(400, "Fail to delete")
		}

		return c.NoContent(200)
	})

	e.GET("/home/edit", func(c echo.Context) error {
		c.Render(200, "home.html/header_buttons_out", nil)
		return c.Render(200, "home.html/block_edit", nil)
	})

	e.POST("/home/edit", func(c echo.Context) error {
		c.Render(200, "home.html/header_buttons", nil)
		return c.NoContent(200)
	})

	e.GET("/home/add/list", func(c echo.Context) error {
		modules := []mod.ModuleMetada{
			{
				Name:  "Letterboxd",
				Icon:  "letterboxd",
				Sizes: []string{"1x1"},
			}, {
				Name:  "Radarr",
				Icon:  "radarr",
				Sizes: []string{"1x1"},
			},
		}

		return c.Render(200, "home.html/add-block-popup", modules)
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
