package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	comp "diikstra.fr/homeboard/components"
	"diikstra.fr/homeboard/models"
	"diikstra.fr/homeboard/pkg/classification"
	database "diikstra.fr/homeboard/pkg/db"
	"diikstra.fr/homeboard/pkg/utils"
	"diikstra.fr/homeboard/services/linkhub"
	"github.com/MathisVerstrepen/go-module/gosoup"
	f "github.com/MathisVerstrepen/go-module/webfetch"
	"github.com/labstack/echo/v4"
)

func LinkHub(c echo.Context) error {
	newPageData := models.PageData{
		Title:      "Linkhub",
		Page:       "linkhub",
		Background: globalPageData.Background,
	}

	linkHubLinks, err := database.DbConn.GetLinkhubLinks()
	if err != nil {
		return err
	}

	fmt.Println(linkHubLinks)

	return Render(c, http.StatusOK, comp.Root(comp.Linkhub(*linkHubLinks), newPageData))
}

func LinkHubPostSite(c echo.Context) error {
	bodyVariables, err := utils.DecodePostBody(c.Request().Body)
	if err != nil {
		return err
	}

	fetcher := f.Fetcher{
		ProxyUrl: "192.168.2.51:1080",
	}

	url := bodyVariables["url"][0]

	domain, err := utils.GetURLDomain(url)
	if err != nil {
		return err
	}

	bodyNode, err := linkhub.ScrapURL(fetcher, url)
	if err != nil {
		return err
	}

	titleNode := gosoup.GetNodeBySelector(bodyNode, &gosoup.HtmlSelector{
		Tag:      "title",
		OnlyTag:  true,
		Multiple: false,
	})
	gosoup.PrintNode(titleNode[0])
	title := gosoup.GetInnerText(titleNode[0])
	fmt.Println(title)

	images := linkhub.ExtractImagesFromBody(fetcher, bodyNode, domain)
	linkhub.LabelizeImages(&images)

	linkHubLink, err := database.DbConn.SetLinkhubLink(database.LinkHubLink{
		Name:    title,
		Url:     url,
		Icon:    "test",
		Is_nsfw: false,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, imageData := range images {
		err := database.DbConn.SetLinkhubImage(database.LinkHubImage{
			Linkhub_id: linkHubLink.Id,
			Ext:        imageData.Ext,
			Image_id:   imageData.Id,
			Is_nsfw:    classification.IsImageNsfw(imageData.Nsfw_classification),
		})

		if err != nil {
			fmt.Println(err)
		}

		err = database.DbConn.SetImageTags(imageData.Id, imageData.Label_classification)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func LinkHubId(c echo.Context) error {
	linkhubId := c.Param("id")
	id, err := strconv.Atoi(linkhubId)
	if err != nil {
		return nil
	}

	linkhub, err := database.DbConn.GetLinkhubLink(id)
	if err != nil {
		return nil
	}

	linkhubImages, err := database.DbConn.GetLinkhubImages(id)
	if err != nil {
		return nil
	}

	newPageData := models.PageData{
		Title:      fmt.Sprintf("Linkhub - %s", linkhub.Name),
		Page:       "linkhub-id",
		Background: globalPageData.Background,
	}

	return Render(c, http.StatusOK, comp.Root(comp.LinkhubId(linkhub, *linkhubImages), newPageData))
}

func LinkHubImageDetails(c echo.Context) error {
	imageId := c.Param("id")

	imageMeta, imageTags, err := database.DbConn.GetImageDetails(imageId)

	if err != nil {
		return err
	}

	return Render(c, http.StatusOK, comp.LinkhubImageDetail(imageMeta, imageTags))
}
