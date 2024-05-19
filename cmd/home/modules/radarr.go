package modules

var radarrModule = Module{
	GetMetadata: func() ModuleMetada {
		return ModuleMetada{Name: "Radarr", Icon: "radarr", Sizes: []string{"1x1"}}
	},
}
