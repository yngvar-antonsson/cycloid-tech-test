package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const openMeteoURI = "https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&start_date=%s&end_date=%s&daily=temperature_2m_max&timezone=UTC"

type OpenMeteo struct{}

type openMeteoResponceType struct {
	Dates []string  `json:"daily.time"`
	Temps []float64 `json:"daily.temperature_2m_max"`
	Daily struct {
		Time        []string  `json:"time"`
		Temperature []float64 `json:"temperature_2m_max"`
	} `json:"daily"`
}

func NewOpenMeteo() *OpenMeteo {
	return &OpenMeteo{}
}

func (o *OpenMeteo) Name() string {
	return "OpenMeteo"
}

func (o *OpenMeteo) GetParams(string) any {
	return nil
}

func (o *OpenMeteo) SetParams(string, any) {
}

func (o *OpenMeteo) GetForecast(ctx context.Context, lat, lon string) (ForecastDay, error) {
	return BuildForecastGetter(ctx, lat, lon, o, openMeteoRequest)(ctx, lat, lon)
}

func openMeteoRequest(ctx context.Context, wg *sync.WaitGroup, lat, lon string, res *sync.Map, i int, o WeatherProvider) {
	defer wg.Done()
	select {
	case <-ctx.Done():
		return
	default:
	}

	currentDate := time.Now().AddDate(0, 0, i).Format("2006-01-02")
	url := fmt.Sprintf(openMeteoURI, lat, lon, currentDate, currentDate)

	resp, err := http.Get(url)
	if err != nil {
		res.Store(currentDate, err)
		return
	}
	defer resp.Body.Close()

	var data openMeteoResponceType

	err = json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		res.Store(currentDate, err)
		return
	}

	res.Store(currentDate, data.Daily.Temperature[0])
}
