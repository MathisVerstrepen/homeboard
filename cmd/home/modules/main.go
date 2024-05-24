package modules

import (
	"errors"

	f "github.com/MathisVerstrepen/go-module/webfetch"
	"github.com/a-h/templ"
	"github.com/redis/go-redis/v9"

	. "diikstra.fr/homeboard/cmd/models"
)

type Module struct {
	GetMetadata func() ModuleMetada
	RenderView  func(*redis.Client, string, string, f.Fetcher) (int, templ.Component, error)
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

func (ms ModuleService) RenderModule(rdb *redis.Client, name string, position string) (int, templ.Component, error) {
	for _, module := range modules {
		if module.GetMetadata().Name == name {
			return module.RenderView(rdb, name, position, (*ms.Proxies)[0])
		}
	}

	return 0, nil, errors.New("no module named " + name + " found")
}
