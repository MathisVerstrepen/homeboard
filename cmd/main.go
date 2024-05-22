package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"

	c "diikstra.fr/homeboard/cmd/cache"
	db "diikstra.fr/homeboard/cmd/database"
	mod "diikstra.fr/homeboard/cmd/home/modules"
	f "github.com/MathisVerstrepen/go-module/webfetch"
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
		"iterate": func(n int) []string {
			return make([]string, n)
		},
		// Create a dict from passed values that can be used in template after
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					continue
				}
				if i+1 < len(values) {
					dict[key] = values[i+1]
				}
			}
			return dict
		},
		"blockIds": func(ncols, nrows int) []string {
			ids := make([]string, ncols*nrows)
			for row := range nrows {
				for col := range ncols {
					ids[row*3+col] = fmt.Sprintf("card_%d_%d", row+1, col+1)
				}
			}
			return ids
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
	Layout *[]db.ModulePosition
}

type HomeAddPopup struct {
	Position string
	Modules  []mod.ModuleMetada
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

	e.Renderer = newTemplate()

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

		return c.Render(200, "home.html", &newPageData)
	})

	e.GET("/home/modules", func(c echo.Context) error {
		modules := moduleService.GetModulesMetadata()
		for _, modulePosition := range *homeLayout.Layout {
			for _, module := range modules {
				if modulePosition.ModuleName == module.Name {
					moduleService.RenderModule(c, cache, module.Name, modulePosition.Position)
				}
			}
		}

		return c.NoContent(200)
	})

	e.POST("/home/modules/:moduleName/:position", func(c echo.Context) error {
		moduleName := c.Param("moduleName")
		position := c.Param("position")

		err := dbConn.SetHomeLayout(position, moduleName)
		if err != nil {
			return err
		}

		homeLayout.Layout = dbConn.GetHomeLayouts()

		moduleService.RenderModule(c, cache, moduleName, position)
		return c.NoContent(200)
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
			Backgrounds: dbConn.GetBackgrounds(),
		})
	})

	e.POST("/settings/backgrounds", func(c echo.Context) error {
		bg, err := dbConn.UploadBackground(c)
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

		background, err = dbConn.SetSelectedBackground(id)
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

		err = dbConn.DeleteBackground(id)
		if err != nil {
			return c.String(400, "Fail to delete")
		}

		return c.NoContent(200)
	})

	e.GET("/home/edit", func(c echo.Context) error {
		c.Render(200, "home.html/header_buttons_out", nil)

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

		return c.Render(200, "home.html/block_edit", ids)
	})

	e.POST("/home/edit", func(c echo.Context) error {
		c.Render(200, "home.html/header_buttons", nil)
		return c.Render(200, "home.html/home_layout", &homeLayout)
	})

	e.GET("/home/add/list/:position", func(c echo.Context) error {
		data := HomeAddPopup{
			Position: c.Param("position"),
			Modules:  moduleService.GetModulesMetadata(),
		}

		return c.Render(200, "home.html/add-block-popup", data)
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
