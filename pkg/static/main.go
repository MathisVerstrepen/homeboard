package static

import (
	"diikstra.fr/homeboard/models"
	f "github.com/MathisVerstrepen/go-module/webfetch"
)

var HomeLayout models.HomeLayoutData
var Proxies *[]f.Fetcher

func SetVariables(moduleName string, modulePosition string, variables map[string]string) {
	for i, moduledata := range HomeLayout.LayoutData {
		if moduledata.Name == moduleName && moduledata.Position == modulePosition {
			HomeLayout.LayoutData[i].Variables = variables
		}
	}
}

func SetProxies(p *[]f.Fetcher) {
	Proxies = p
}
