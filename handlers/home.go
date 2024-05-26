package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	comp "diikstra.fr/homeboard/components"
	"diikstra.fr/homeboard/models"
)

func HomeHandler(c echo.Context) error {
	newPageData := models.PageData{
		Title:      "Home",
		Page:       "home",
		Background: globalPageData.Background,
		HomeLayout: homeLayout,
	}

	return Render(c, http.StatusOK, comp.Root(comp.Home(homeLayout), newPageData))
}

func HomeModulesHandler(c echo.Context) error {
	modules := moduleService.GetModulesMetadata()
	for _, modulePosition := range *homeLayout.Layout {
		for _, module := range modules {
			if modulePosition.ModuleName == module.Name {
				statusCode, component, error := moduleService.RenderModule(cache, module.Name, modulePosition.Position)

				if error != nil {
					Render(c, http.StatusBadRequest, nil)
				} else {
					Render(c, statusCode, component)
				}
			}
		}
	}

	return nil
}

func HomeAddModulePositionHandler(c echo.Context) error {
	moduleName := c.Param("moduleName")
	position := c.Param("position")

	err := dbConn.SetHomeLayout(position, moduleName)
	if err != nil {
		return err
	}

	homeLayout.Layout = dbConn.GetHomeLayouts()

	statusCode, component, error := moduleService.RenderModule(cache, moduleName, position)
	if error != nil {
		return Render(c, http.StatusBadRequest, nil)
	} else {
		return Render(c, statusCode, component)
	}
}

func HomeGetEdit(c echo.Context) error {
	Render(c, http.StatusOK, comp.Header_buttons_out())

	nCols := homeLayout.NCols
	nRows := homeLayout.NRows
	ids := make([]string, 0)
	for row := range nRows {
		for col := range nCols {
			position := fmt.Sprintf("card_%d_%d", row+1, col+1)

			present := false
			for _, layout := range *homeLayout.Layout {
				if layout.Position == position {
					present = true
					break
				}
			}

			if !present {
				ids = append(ids, position)
			}
		}
	}

	return Render(c, http.StatusOK, comp.BlockEdit(ids))
}

func HomePostEdit(c echo.Context) error {
	Render(c, http.StatusOK, comp.Header_buttons())
	return Render(c, http.StatusOK, comp.HomeLayout(homeLayout))
}

func HomeGetAddList(c echo.Context) error {
	addPopupData := models.HomeAddPopup{
		Position: c.Param("position"),
		Modules:  moduleService.GetModulesMetadata(),
	}

	return Render(c, http.StatusOK, comp.AddBlockPopup(addPopupData))
}
