package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type WeatherAPI struct {
	params map[string]any
}

func NewWeatherAPI() *WeatherAPI {
	return &WeatherAPI{
		params: make(map[string]any),
	}
}

func (w *WeatherAPI) Name() string {
	return "WeatherAPI"
}

func (w *WeatherAPI) GetParams(name string) any {
	return w.params[name]
}

func (w *WeatherAPI) SetParams(name string, value any) {
	w.params[name] = value
}

func (w *WeatherAPI) GetForecast(ctx context.Context, lat, lon string) (ForecastDay, error) {
	return BuildForecastGetter(ctx, lat, lon, w, weatherAPIRrequest)(ctx, lat, lon)
}

const weatherAPIURI = "https://api.weatherapi.com/v1/forecast.json?key=%s&q=%s,%s&dt=%s"

type weatherAPIResponceType struct {
	Forecast struct {
		Forecastday []struct {
			Date string `json:"date"`
			Day  struct {
				MaxTempC float64 `json:"maxtemp_c"`
			} `json:"day"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func weatherAPIRrequest(ctx context.Context, wg *sync.WaitGroup, lat, lon string, res *sync.Map, i int, wp WeatherProvider) {
	defer wg.Done()
	select {
	case <-ctx.Done():
		return
	default:
	}
	currentDate := time.Now().AddDate(0, 0, i).Format("2006-01-02")
	url := fmt.Sprintf(weatherAPIURI, wp.GetParams("APIKey").(string), lat, lon, currentDate)

	resp, err := http.Get(url)
	if err != nil {
		res.Store(currentDate, err)
		return
	}
	defer resp.Body.Close()

	var data weatherAPIResponceType

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		res.Store(currentDate, err)
		return
	}
	res.Store(currentDate, data.Forecast.Forecastday[0].Day.MaxTempC)
}
