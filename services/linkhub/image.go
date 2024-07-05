package linkhub

import (
	"strings"

	"diikstra.fr/homeboard/pkg/classification"
	"github.com/MathisVerstrepen/go-module/gosoup"
	f "github.com/MathisVerstrepen/go-module/webfetch"
	"golang.org/x/net/html"
)

type Image struct {
	Url                  string
	Id                   string
	Ext                  string
	Path                 string
	Nsfw_classification  []classification.Classification
	Label_classification []classification.Classification
}

func ExtractImagesFromBody(fetcher f.Fetcher, bodyNode *html.Node, domain string) []Image {
	imgNodes := gosoup.GetNodeBySelector(bodyNode, &gosoup.HtmlSelector{
		Tag:     "img",
		OnlyTag: true,
	})

	urls := []Image{}

	for _, imgNode := range imgNodes {
		imgSrc := gosoup.GetAttribute(imgNode, "src")

		// Some site store their images in base64 format in src attribute
		// In this case image link is likely to be in data-src attribute
		if strings.Contains(imgSrc, "base64") {
			imgSrc = gosoup.GetAttribute(imgNode, "data-src")
		}

		if len(imgSrc) == 0 {
			continue
		}

		// Some images doesn't include domain (relative url)
		if string(imgSrc[0]) == "/" {
			imgSrc = "https://" + domain + imgSrc
		}

		path, id, ext, err := DownloadImage(fetcher, imgSrc)
		if err != nil {
			continue
		}

		urls = append(urls, Image{
			Id:   id,
			Ext:  ext,
			Url:  imgSrc,
			Path: path,
		})
	}

	return urls
}

func LabelizeImages(urls *[]Image) {
	for i, urlData := range *urls {
		labels_nsfw, err := classification.LabelizeImageNSFW(urlData.Path)
		if err != nil {
			continue
		}

		labels, err := classification.LabelizeImage(urlData.Path)
		if err != nil {
			continue
		}

		(*urls)[i].Nsfw_classification = labels_nsfw
		(*urls)[i].Label_classification = labels
	}
}
