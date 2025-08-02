package handler

import (
	"cycloid/test/provider"
)

type Settings struct {
	Providers []provider.WeatherProvider
	APILimit  int
}

var settings Settings

func setParams(provider provider.WeatherProvider, params map[string]any) {
	for name, param := range params {
		provider.SetParams(name, param)
	}
}

func Setup(apiLimit int, config map[string]map[string]any) {
	for providerName, providerSettings := range config {
		switch providerName {
		case "openmeteo":
			openmeteo := provider.NewOpenMeteo()
			setParams(openmeteo, providerSettings)
			settings.Providers = append(settings.Providers, openmeteo)
		case "weatherapi":
			weatherapi := provider.NewWeatherAPI()
			setParams(weatherapi, providerSettings)
			settings.Providers = append(settings.Providers, weatherapi)
		}
	}
	settings.APILimit = apiLimit
}
