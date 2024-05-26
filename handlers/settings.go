package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	comp "diikstra.fr/homeboard/components"
	"diikstra.fr/homeboard/models"
)

func SettingsHandler(c echo.Context) error {
	newPageData := models.PageData{
		Title:      "Settings",
		Page:       "settings",
		Background: globalPageData.Background,
	}

	return Render(c, http.StatusOK, comp.Root(comp.Settings(), newPageData))
}

func SettingsGetBackgrounds(c echo.Context) error {
	return Render(c, http.StatusOK, comp.BgPopup(models.BackgroundData{
		Backgrounds: dbConn.GetBackgrounds(),
	}))
}

func SettingsPostBackground(c echo.Context) error {
	bg, err := dbConn.UploadBackground(c)
	if err != nil {
		return err
	}

	return Render(c, http.StatusOK, comp.BgItem(bg))
}

func SettingsSetSelectedBackgroundfunc(c echo.Context) error {
	idBg := c.Param("id")
	id, err := strconv.Atoi(idBg)
	if err != nil {
		return c.String(400, "Invalid id")
	}

	Render(c, http.StatusOK, comp.OobButtonBgSelect(globalPageData.Background))

	background, err := dbConn.SetSelectedBackground(id)
	if err != nil {
		return c.String(400, "Fail to set new background :"+err.Error())
	}
	globalPageData.Background = background

	Render(c, http.StatusOK, comp.OobButtonBgSelected(background))
	return Render(c, http.StatusOK, comp.Background(globalPageData))
}

func SettingsDeleteBackground(c echo.Context) error {
	idBg := c.Param("id")
	id, err := strconv.Atoi(idBg)
	if err != nil {
		return nil
	}

	err = dbConn.DeleteBackground(id)
	if err != nil {
		return nil
	}

	return nil
}
