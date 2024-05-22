package modules

import (
	"log"
	"os"
	"strings"

	c "diikstra.fr/homeboard/cmd/cache"
	gs "github.com/MathisVerstrepen/go-module/gosoup"
	f "github.com/MathisVerstrepen/go-module/webfetch"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"golang.org/x/net/html"
)

type MovieData struct {
	Poster      string
	Owner       string
	OwnerAvatar string
	Rating      string
	Slug        string
	Id          string
}

type RenderData struct {
	MovieData []MovieData
	Metadata  ModuleMetada
}

var letterboxdModuleMetada = ModuleMetada{
	Name:     "Letterboxd",
	Icon:     "letterboxd",
	Sizes:    []string{"1x1"},
	Position: "",
	CacheKey: "letterboxd_recent__friends_movies",
}

var letterboxdModule = Module{
	GetMetadata: func() ModuleMetada {
		return letterboxdModuleMetada
	},
	RenderView: func(ctx echo.Context, rdb *redis.Client, name string, position string, fetcher f.Fetcher) {
		var moviesData []MovieData

		err := c.GetCachedKey(rdb, letterboxdModuleMetada.CacheKey, &moviesData)
		if err != nil {
			moviesData = GetFriendsRecentMovies(fetcher)
			err := c.SetCachedKey(rdb, letterboxdModuleMetada.CacheKey, moviesData)
			if err != nil {
				log.Printf("fail to set key %s in cache", letterboxdModuleMetada.CacheKey)
				log.Printf("%v", err)
			}
		}

		letterboxdModuleMetada.Position = position
		ctx.Render(200, "letterboxd.html/card", &RenderData{
			MovieData: moviesData,
			Metadata:  letterboxdModuleMetada,
		})
	},
}

func GetFriendsRecentMovies(fetcher f.Fetcher) []MovieData {
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
		return []MovieData{}
	}

	friendsMovies := gs.GetNodeByClass(groupNode[0], &gs.HtmlSelector{
		Tag:        "li",
		ClassNames: "poster-container viewing-poster-container",
		Multiple:   true,
	})

	var moviesData []MovieData
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

		moviesData = append(moviesData, MovieData{
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
