package modules

import (
	"log"
	"net/http"
	"os"
	"strings"

	comp "diikstra.fr/homeboard/components"
	c "diikstra.fr/homeboard/pkg/cache"
	gs "github.com/MathisVerstrepen/go-module/gosoup"
	f "github.com/MathisVerstrepen/go-module/webfetch"
	"github.com/a-h/templ"
	"github.com/redis/go-redis/v9"

	"diikstra.fr/homeboard/models"

	"golang.org/x/net/html"
)

var letterboxdModuleMetada = models.ModuleMetada{
	Name:  "Letterboxd",
	Icon:  "letterboxd",
	Sizes: []string{"1x1"},
}

var letterboxdModule = models.Module{
	GetMetadata: func() models.ModuleMetada {
		return letterboxdModuleMetada
	},
	RenderView: func(rdb *redis.Client, name string, position string, fetcher f.Fetcher, useCache bool) (int, templ.Component, error) {
		return http.StatusOK, comp.LetterboxdCard(
			renderLetterboxdDataConstructor(rdb, name, position, fetcher, useCache),
		), nil
	},
	RenderViewContent: func(rdb *redis.Client, name string, position string, fetcher f.Fetcher, useCache bool) (int, templ.Component, error) {
		return http.StatusOK, comp.LetterboxdCardContent(
			renderLetterboxdDataConstructor(rdb, name, position, fetcher, useCache),
		), nil
	},
}

func renderLetterboxdDataConstructor(rdb *redis.Client, name string, position string, fetcher f.Fetcher, useCache bool) models.LetterboxdRenderData {
	var moviesData []models.LetterboxdMovieData

	moduleData := GetModuleData(name, position)

	err := c.GetCachedKey(rdb, moduleData.CacheKey, &moviesData)
	if err != nil || !useCache {
		moviesData = GetFriendsRecentMovies(fetcher)
		err := c.SetCachedKey(rdb, moduleData.CacheKey, moviesData)
		if err != nil {
			log.Printf("fail to set key %s in cache", moduleData.CacheKey)
			log.Printf("%v", err)
		}
	}

	return models.LetterboxdRenderData{
		MovieData: moviesData,
		Metadata:  letterboxdModuleMetada,
		Data:      moduleData,
	}
}

func GetFriendsRecentMovies(fetcher f.Fetcher) []models.LetterboxdMovieData {
	body := fetcher.FetchData(f.FetcherParams{
		Method: "GET",
		Url:    "https://letterboxd.com/",
		Body:   nil,
		Headers: f.Header{
			"Cookie": os.Getenv("LETTERBOXD_COOKIES"),
		},
		Params:       f.Param{},
		UseProxy:     true,
		WantErrCodes: nil,
	})

	parsedBody, _ := html.Parse(strings.NewReader(string(body)))

	groupNode := gs.GetNodeByClass(parsedBody, &gs.HtmlSelector{
		Id:         "recent-from-friends",
		Tag:        "section",
		ClassNames: "",
		Multiple:   false,
	})

	if len(groupNode) == 0 {
		log.Println("Failed to get recent friends watch")
		return []models.LetterboxdMovieData{}
	}

	friendsMovies := gs.GetNodeByClass(groupNode[0], &gs.HtmlSelector{
		Tag:        "li",
		ClassNames: "poster-container viewing-poster-container",
		Multiple:   true,
	})

	var moviesData []models.LetterboxdMovieData
	for _, movieNode := range friendsMovies[:len(friendsMovies)-1] {
		owner := gs.GetAttribute(movieNode, "data-owner")
		filmId := gs.GetAttribute(movieNode.FirstChild.NextSibling, "data-film-id")
		filmSlug := gs.GetAttribute(movieNode.FirstChild.NextSibling, "data-film-slug")

		posterUrl := getPosterURL(fetcher, filmSlug)

		ownerImgNode := gs.GetNodeByClass(movieNode, &gs.HtmlSelector{
			Tag:        "a",
			ClassNames: "avatar -a16",
			Multiple:   false,
		})
		ownerPfpUrl := gs.GetAttribute(ownerImgNode[0].FirstChild.NextSibling, "src")

		ratingNode := gs.GetNodeByClass(movieNode, &gs.HtmlSelector{
			Tag:        "span",
			ClassNames: "rating",
			Multiple:   false,
		})
		rating := "\u2001"
		if len(ratingNode) > 0 {
			rating = strings.TrimSpace(ratingNode[0].FirstChild.Data)
		}

		moviesData = append(moviesData, models.LetterboxdMovieData{
			Poster:      posterUrl,
			Owner:       owner,
			OwnerAvatar: ownerPfpUrl,
			Rating:      rating,
			Slug:        filmSlug,
			Id:          filmId,
		})
	}

	return moviesData
}

func getPosterURL(fetcher f.Fetcher, filmSlug string) string {
	body := fetcher.FetchData(f.FetcherParams{
		Method: "GET",
		Url:    "https://letterboxd.com/ajax/poster/film/" + filmSlug + "/std/150x225/",
		Body:   nil,
		Headers: f.Header{
			"Cookie": os.Getenv("LETTERBOXD_COOKIES"),
		},
		Params:       f.Param{},
		UseProxy:     true,
		WantErrCodes: nil,
	})

	parsedBody, _ := html.Parse(strings.NewReader(string(body)))

	imgNode := gs.GetNodeByClass(parsedBody, &gs.HtmlSelector{
		Id:         "",
		Tag:        "img",
		ClassNames: "image",
		Multiple:   false,
	})

	return gs.GetAttribute(imgNode[0], "src")
}
