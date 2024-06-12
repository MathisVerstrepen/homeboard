package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	comp "diikstra.fr/homeboard/components"
	"diikstra.fr/homeboard/models"
	database "diikstra.fr/homeboard/pkg/db"
	"diikstra.fr/homeboard/pkg/static"
	"diikstra.fr/homeboard/services/home/modules"
)

func HomeHandler(c echo.Context) error {
	newPageData := models.PageData{
		Title:      "Home",
		Page:       "home",
		Background: globalPageData.Background,
		HomeLayout: static.HomeLayout,
	}

	return Render(c, http.StatusOK, comp.Root(comp.Home(static.HomeLayout), newPageData))
}

func HomeGetEdit(c echo.Context) error {
	Render(c, http.StatusOK, comp.Header_buttons_out())

	nCols := static.HomeLayout.NCols
	nRows := static.HomeLayout.NRows
	ids := make([]string, 0)
	for row := range nRows {
		for col := range nCols {
			position := fmt.Sprintf("card_%d_%d", row+1, col+1)

			present := false
			for _, moduleData := range static.HomeLayout.LayoutData {
				if moduleData.Position == position {
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
	return Render(c, http.StatusOK, comp.HomeLayout(static.HomeLayout))
}

func HomeGetAddList(c echo.Context) error {
	addPopupData := models.HomeAddPopup{
		Position: c.Param("position"),
		Modules:  moduleService.GetModulesMetadata(),
	}

	return Render(c, http.StatusOK, comp.AddBlockPopup(addPopupData))
}

func handleModuleRender(c echo.Context, moduleName string, position string, useCache bool) error {
	modules := moduleService.GetModulesMetadata()
	for _, module := range modules {
		if moduleName == module.Name {
			statusCode, component, error := moduleService.RenderModule(cache, module.Name, position, useCache)

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

func HomeModuleHandler(c echo.Context) error {
	moduleName := c.Param("moduleName")
	position := c.Param("position")

	return handleModuleRender(c, moduleName, position, true)
}

func HomeModuleHandlerForceRefresh(c echo.Context) error {
	moduleName := c.Param("moduleName")
	position := c.Param("position")

	return handleModuleRender(c, moduleName, position, false)
}

func HomeModulesHandler(c echo.Context) error {
	modules := moduleService.GetModulesMetadata()
	for _, moduleData := range static.HomeLayout.LayoutData {
		for _, module := range modules {
			if moduleData.Name == module.Name {
				statusCode, component, error := moduleService.RenderModule(cache, module.Name, moduleData.Position, true)

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
		return err
	}

	newLayoutData := make([]models.ModuleData, len(static.HomeLayout.LayoutData)-1)
	for i, moduleData := range static.HomeLayout.LayoutData {
		if moduleData.Position == position {
			copy(newLayoutData[i:len(static.HomeLayout.LayoutData)-1], static.HomeLayout.LayoutData[i+1:len(static.HomeLayout.LayoutData)])
			break
		}
		newLayoutData[i] = moduleData
	}
	static.HomeLayout.LayoutData = newLayoutData

	return Render(c, http.StatusOK,
		comp.Alert("success", "Suppression effectuée", fmt.Sprintf("Suppresion du module %s effectuée avec succès !", moduleName)))
}

func HomeAddModulePositionHandler(c echo.Context) error {
	moduleName := c.Param("moduleName")
	position := c.Param("position")

	err := database.DbConn.SetHomeLayout(position, moduleName)
	if err != nil {
		return err
	}

	moduleMetadata, err := modules.GetModuleMetadata(moduleName)
	if err != nil {
		return err
	}

	static.HomeLayout.LayoutData = append(static.HomeLayout.LayoutData, models.ModuleData{
		Name:      moduleName,
		Position:  position,
		CacheKey:  fmt.Sprintf("module_%s_%s", moduleName, position),
		Variables: moduleMetadata.DefaultVariables,
	})

	statusCode, component, error := moduleService.RenderModule(cache, moduleName, position, true)
	if error != nil {
		return Render(c, http.StatusBadRequest, nil)
	}
	return Render(c, statusCode, component)
}

func HomeGetModuleEdit(c echo.Context) error {
	moduleName := c.Param("moduleName")
	position := c.Param("position")

	return Render(c, http.StatusOK, comp.ModuleEdit(moduleName, position))
}

func HomeGetModuleEditVariables(c echo.Context) error {
	moduleName := c.Param("moduleName")
	position := c.Param("position")

	moduleMetadata, err := modules.GetModuleMetadata(moduleName)
	moduleData := modules.GetModuleData(moduleName, position)
	if err != nil {
		return err
	}

	return Render(c, http.StatusOK, comp.ModuleEditVariables(moduleMetadata, moduleData))
}

func HomePostModuleEditVariables(c echo.Context) error {
	moduleName := c.Param("moduleName")
	position := c.Param("position")

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	variablesUrlValues, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}

	// url.Values if a map of string to list of string so we need to cast to a standard map string to string
	variables := make(map[string]string)
	for variableName, variableValue := range variablesUrlValues {
		variables[variableName] = variableValue[0]
	}

	oldVariables := make(map[string]string)
	database.DbConn.GetModuleVariables(position, &oldVariables)

	err = database.DbConn.SetModuleVariables(position, &variables)
	if err != nil {
		return err
	}

	variablesChanged := false
	for variableName, variableValue := range variables {
		if variableValue != oldVariables[variableName] {
			variablesChanged = true
			break
		}
	}

	if variablesChanged {
		static.SetVariables(moduleName, position, variables)
		statusCode, component, error := moduleService.RenderModuleContent(cache, moduleName, position, false)
		if error != nil {
			return Render(c, http.StatusBadRequest, nil)
		} else {
			Render(c, http.StatusOK,
				comp.Alert("success", "Enregistrement effectué", "Enregistrement des variables effectué avec succès !"))
			return Render(c, statusCode, component)
		}
	}

	return nil
}
