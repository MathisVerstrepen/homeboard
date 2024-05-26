package handlers

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	f "github.com/MathisVerstrepen/go-module/webfetch"
	"github.com/a-h/templ"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"diikstra.fr/homeboard/models"
	c "diikstra.fr/homeboard/pkg/cache"
	database "diikstra.fr/homeboard/pkg/db"
	mod "diikstra.fr/homeboard/services/home/modules"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

var dbConn *database.Conn
var cache *redis.Client

var homeLayout models.HomeLayoutData
var proxies *[]f.Fetcher
var moduleService mod.ModuleService
var globalPageData models.PageData

func Init() {
	fmt.Println("Startup sequence starting...")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbConn = database.Connect()
	cache = c.Connect()

	homeLayout = models.HomeLayoutData{
		NRows:  3,
		NCols:  3,
		Layout: dbConn.GetHomeLayouts(),
	}

	proxies = f.InitFetchers(filepath.Join(basepath, ".."))
	moduleService = mod.ModuleService{
		Proxies: proxies,
	}

	background, err := dbConn.GetSelectedBackground()
	if err != nil {
		log.Printf("%v", err)
		log.Fatal("Fail to fetch selected background")
	}
	globalPageData = models.PageData{
		Background: background,
	}

	fmt.Println("Startup sequence done.")
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}
