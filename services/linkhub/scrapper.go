package linkhub

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"

	f "github.com/MathisVerstrepen/go-module/webfetch"
	"golang.org/x/net/html"
)

func ScrapURL(fetcher f.Fetcher, url string) (*html.Node, error) {
	body, err := Run(context.Background(), url, false, 2*time.Minute)

	if err != nil {
		return nil, err
	}

	return html.Parse(strings.NewReader(*body))
}

func DownloadImage(fetcher f.Fetcher, imageUrl string) (string, string, string, error) {
	urlSep := strings.Split(imageUrl, ".")
	ext := urlSep[len(urlSep)-1]

	if strings.Contains(imageUrl, ".svg") || strings.Contains(imageUrl, ".gif") {
		return "", "", "", errors.New("unwanted format")
	}

	imgByte, err := fetcher.FetchData(f.FetcherParams{
		Method:       "GET",
		Url:          imageUrl,
		Body:         nil,
		Headers:      f.Header{},
		Params:       f.Param{},
		UseProxy:     true,
		WantErrCodes: []int{200},
	})

	if err != nil {
		fmt.Println(err)
		return "", "", "", err
	}

	id := uuid.Must(uuid.NewV4()).String()
	path := fmt.Sprintf("assets/images/linkhub/%s.%s", id, ext)
	out, _ := os.Create(path)
	defer out.Close()

	err = os.WriteFile(path, imgByte, 0644)

	return path, id, ext, err
}
