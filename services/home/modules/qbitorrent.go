package modules

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	comp "diikstra.fr/homeboard/components"
	c "diikstra.fr/homeboard/pkg/cache"
	database "diikstra.fr/homeboard/pkg/db"
	f "github.com/MathisVerstrepen/go-module/webfetch"
	"github.com/a-h/templ"
	"github.com/redis/go-redis/v9"

	"diikstra.fr/homeboard/models"
)

var qbitorrentModuleMetada = models.ModuleMetada{
	Name:     "Qbitorrent",
	Icon:     "qbitorrent",
	Sizes:    []string{"1x1"},
	Position: "",
	CacheKey: "qbitorrent_global",
	Variables: map[string]string{
		"Host": "192.168.2.64",
		"Port": "8114",
	},
}

var qbitorrentModule = models.Module{
	GetMetadata: func() models.ModuleMetada {
		return qbitorrentModuleMetada
	},
	RenderView: func(rdb *redis.Client, name string, position string, fetcher f.Fetcher, useCache bool) (int, templ.Component, error) {
		return http.StatusOK, comp.QbitorrentCard(
			renderQbittorrentDataConstructor(rdb, position, fetcher, useCache),
		), nil
	},
	RenderViewContent: func(rdb *redis.Client, name string, position string, fetcher f.Fetcher, useCache bool) (int, templ.Component, error) {
		return http.StatusOK, comp.QbitorrentCardContent(
			renderQbittorrentDataConstructor(rdb, position, fetcher, useCache),
		), nil
	},
}

func renderQbittorrentDataConstructor(rdb *redis.Client, position string, fetcher f.Fetcher, useCache bool) models.QbitorrentRenderData {
	var qbittorentData models.QbitorrentGlobalData

	qbitorrentModuleMetada.Position = position
	database.DbConn.GetModuleVariables(position, &qbitorrentModuleMetada.Variables)

	err := c.GetCachedKey(rdb, qbitorrentModuleMetada.CacheKey, &qbittorentData)
	if err != nil || !useCache {
		qbittorentData = GetQbittorentGlobalData(fetcher, qbitorrentModuleMetada.Variables["Host"], qbitorrentModuleMetada.Variables["Port"])
		err := c.SetCachedKey(rdb, qbitorrentModuleMetada.CacheKey, qbittorentData)
		if err != nil {
			log.Printf("fail to set key %s in cache", qbitorrentModuleMetada.CacheKey)
			log.Printf("%v", err)
		}
	}

	return models.QbitorrentRenderData{
		QbitorrentGlobalData: qbittorentData,
		Metadata:             qbitorrentModuleMetada,
	}
}

func GetQbittorentGlobalData(fetcher f.Fetcher, host string, port string) models.QbitorrentGlobalData {
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

	var qbittorentData models.QbitorrentGlobalData
	json.Unmarshal(body, &qbittorentData)

	return qbittorentData
}
