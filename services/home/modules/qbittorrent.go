package modules

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

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
		"Host":     "192.168.2.64",
		"Port":     "8114",
		"Username": "",
		"Password": "",
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
		var authCookie string
		if moduleData.Variables["Username"] != "" && moduleData.Variables["Password"] != "" {
			authCookie, _ = GetQbittorentAuthCookies(fetcher, moduleData.Variables["Host"], moduleData.Variables["Port"], moduleData.Variables["Username"], moduleData.Variables["Password"])
		}

		qbittorentData = GetQbittorentGlobalData(fetcher, moduleData.Variables["Host"], moduleData.Variables["Port"], authCookie)
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

func GetQbittorentGlobalData(fetcher f.Fetcher, host string, port string, authCookie string) models.QbittorrentGlobalData {
	body, _ := fetcher.FetchData(f.FetcherParams{
		Method: "GET",
		Url:    fmt.Sprintf("http://%s:%s/api/v2/sync/maindata", host, port),
		Body:   nil,
		Headers: f.Header{
			"Accept": "application/json",
			"Cookie": authCookie,
		},
		Params:       f.Param{},
		UseProxy:     false,
		WantErrCodes: nil,
	})

	var qbittorentData models.QbittorrentGlobalData
	json.Unmarshal(body, &qbittorentData)

	return qbittorentData
}

func GetQbittorentAuthCookies(fetcher f.Fetcher, host string, port string, username string, password string) (string, error) {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	encodedData := data.Encode()

	body, cookies, err := fetcher.FetchDataAndCookies(f.FetcherParams{
		Method: "POST",
		Url:    fmt.Sprintf("http://%s:%s/api/v2/auth/login", host, port),
		Body:   encodedData,
		Headers: f.Header{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Params:       f.Param{},
		UseProxy:     false,
		WantErrCodes: []int{200},
	})

	if string(body) != "Ok." {
		return "", errors.New("authentification failed")
	}

	if err != nil {
		return "", err
	}

	return cookies[0].Raw, nil
}
