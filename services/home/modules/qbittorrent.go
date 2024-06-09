package modules

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	comp "diikstra.fr/homeboard/components"
	c "diikstra.fr/homeboard/pkg/cache"
	f "github.com/MathisVerstrepen/go-module/webfetch"
	"github.com/a-h/templ"
	"github.com/redis/go-redis/v9"

	"diikstra.fr/homeboard/models"
)

var qbittorrentModuleMetada = models.ModuleMetada{
	Name:  "Qbittorrent",
	Icon:  "qbittorrent",
	Sizes: []string{"1x1"},
	DefaultVariables: map[string]string{
		"Host": "192.168.2.64",
		"Port": "8114",
	},
}

var qbittorrentModule = models.Module{
	GetMetadata: func() models.ModuleMetada {
		return qbittorrentModuleMetada
	},
	RenderView: func(rdb *redis.Client, name string, position string, fetcher f.Fetcher, useCache bool) (int, templ.Component, error) {
		return http.StatusOK, comp.QbittorrentCard(
			renderQbittorrentDataConstructor(rdb, name, position, fetcher, useCache),
		), nil
	},
	RenderViewContent: func(rdb *redis.Client, name string, position string, fetcher f.Fetcher, useCache bool) (int, templ.Component, error) {
		return http.StatusOK, comp.QbittorrentCardContent(
			renderQbittorrentDataConstructor(rdb, name, position, fetcher, useCache),
		), nil
	},
}

func renderQbittorrentDataConstructor(rdb *redis.Client, name string, position string, fetcher f.Fetcher, useCache bool) models.QbittorrentRenderData {
	var qbittorentData models.QbittorrentGlobalData

	moduleData := GetModuleData(name, position)

	err := c.GetCachedKey(rdb, moduleData.CacheKey, &qbittorentData)
	if err != nil || !useCache {
		qbittorentData = GetQbittorentGlobalData(fetcher, moduleData.Variables["Host"], moduleData.Variables["Port"])
		err := c.SetCachedKey(rdb, moduleData.CacheKey, qbittorentData)
		if err != nil {
			log.Printf("fail to set key %s in cache", moduleData.CacheKey)
			log.Printf("%v", err)
		}
	}

	return models.QbittorrentRenderData{
		QbittorrentGlobalData: qbittorentData,
		Metadata:              qbittorrentModuleMetada,
		Data:                  moduleData,
	}
}

func GetQbittorentGlobalData(fetcher f.Fetcher, host string, port string) models.QbittorrentGlobalData {
	body := fetcher.FetchData(f.FetcherParams{
		Method: "GET",
		Url:    fmt.Sprintf("http://%s:%s/api/v2/sync/maindata", host, port),
		Body:   nil,
		Headers: f.Header{
			"Accept": "application/json",
		},
		Params:       f.Param{},
		UseProxy:     false,
		WantErrCodes: nil,
	})

	var qbittorentData models.QbittorrentGlobalData
	json.Unmarshal(body, &qbittorentData)

	return qbittorentData
}
