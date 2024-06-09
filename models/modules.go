package models

import (
	f "github.com/MathisVerstrepen/go-module/webfetch"
	"github.com/a-h/templ"
	"github.com/redis/go-redis/v9"
)

type ModuleService struct {
	Proxies *[]f.Fetcher
}

type ModuleMetada struct {
	Name             string
	Icon             string
	Sizes            []string
	DefaultVariables map[string]string
}

type ModuleData struct {
	Name      string
	Position  string
	CacheKey  string
	Variables map[string]string
}

type Module struct {
	GetMetadata       func() ModuleMetada
	RenderView        func(*redis.Client, string, string, f.Fetcher, bool) (int, templ.Component, error)
	RenderViewContent func(*redis.Client, string, string, f.Fetcher, bool) (int, templ.Component, error)
}
