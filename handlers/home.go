package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	comp "diikstra.fr/homeboard/components"
	"diikstra.fr/homeboard/models"
	database "diikstra.fr/homeboard/pkg/db"
	"diikstra.fr/homeboard/services/home/modules"
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

func HomeModuleHandler(c echo.Context) error {
	moduleName := c.Param("moduleName")
	position := c.Param("position")

	modules := moduleService.GetModulesMetadata()
	for _, module := range modules {
		if moduleName == module.Name {
			statusCode, component, error := moduleService.RenderModule(cache, module.Name, position)

			if error != nil {
				return error
			} else {
				Render(c, statusCode, component)
			}
			return nil
		}
	}
	return nil
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

func HomeModuleDelete(c echo.Context) error {
	moduleName := c.Param("moduleName")
	position := c.Param("position")

	err := database.DbConn.DeleteHomeLayout(position, moduleName)
	if err != nil {
		fmt.Println(err)
		return err
	}

	homeLayout.Layout = database.DbConn.GetHomeLayouts()

	return nil
}

func HomeAddModulePositionHandler(c echo.Context) error {
	moduleName := c.Param("moduleName")
	position := c.Param("position")

	err := database.DbConn.SetHomeLayout(position, moduleName)
	if err != nil {
		return err
	}

	homeLayout.Layout = database.DbConn.GetHomeLayouts()

	statusCode, component, error := moduleService.RenderModule(cache, moduleName, position)
	if error != nil {
		return Render(c, http.StatusBadRequest, nil)
	} else {
		return Render(c, statusCode, component)
	}
}

func HomeGetModuleEdit(c echo.Context) error {
	moduleName := c.Param("moduleName")
	position := c.Param("position")

	return Render(c, http.StatusOK, comp.ModuleEdit(moduleName, position))
}

func HomeGetModuleEditVariables(c echo.Context) error {
	moduleName := c.Param("moduleName")

	fmt.Println(moduleName)
	moduleMetadata, err := modules.GetModuleMetadata(moduleName)
	fmt.Println(moduleMetadata)
	if err != nil {
		return err
	}

	return Render(c, http.StatusOK, comp.ModuleEditVariables(moduleMetadata))
}
