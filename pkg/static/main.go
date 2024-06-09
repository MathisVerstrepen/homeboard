package static

import (
	"diikstra.fr/homeboard/models"
)

var HomeLayout models.HomeLayoutData

func SetVariables(moduleName string, modulePosition string, variables map[string]string) {
	for i, moduledata := range HomeLayout.LayoutData {
		if moduledata.Name == moduleName && moduledata.Position == modulePosition {
			HomeLayout.LayoutData[i].Variables = variables
		}
	}
}
