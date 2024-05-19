package modules

import (
	"errors"

	f "github.com/MathisVerstrepen/go-module/webfetch"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type ModuleMetada struct {
	Name     string
	Icon     string
	Sizes    []string
	Position string
	CacheKey string
}

type Module struct {
	GetMetadata func() ModuleMetada
	RenderView  func(echo.Context, *redis.Client, string, f.Fetcher)
}

var modules = []Module{
	letterboxdModule,
	// radarrModule,
}

type ModuleService struct {
	Proxies *[]f.Fetcher
}

func (ms ModuleService) GetModulesMetadata() []ModuleMetada {
	var modulesMetadata []ModuleMetada
	for _, module := range modules {
		modulesMetadata = append(modulesMetadata, module.GetMetadata())
	}
	return modulesMetadata
}

func (ms ModuleService) RenderModule(ctx echo.Context, rdb *redis.Client, name string) error {
	for _, module := range modules {
		if module.GetMetadata().Name == name {
			module.RenderView(ctx, rdb, name, (*ms.Proxies)[0])
			return nil
		}
	}

	return errors.New("feur")
}
