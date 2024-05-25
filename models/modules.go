package models

import (
	f "github.com/MathisVerstrepen/go-module/webfetch"
	"github.com/a-h/templ"
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
	RenderView  func(*redis.Client, string, string, f.Fetcher) (int, templ.Component, error)
}
