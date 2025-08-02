package provider

import (
	"context"
	"fmt"
	"sync"
)

type ForecastData struct {
	Temperature float64 `json:"temperature"`
}

type ForecastDay map[string]ForecastData

type ProviderForecast map[string]ForecastDay

const FetchDaysCount = 5

type WeatherProvider interface {
	GetParams(string) any
	SetParams(string, any)
	Name() string
	GetForecast(ctx context.Context, lat, lon string) (ForecastDay, error)
}

type RequestFunc func(context.Context, *sync.WaitGroup, string, string, *sync.Map, int, WeatherProvider)

func BuildForecastGetter(ctx context.Context, lat, lon string, wp WeatherProvider, requestFunc RequestFunc) func(ctx context.Context, lat, lon string) (ForecastDay, error) {
	return func(ctx context.Context, lat, lon string) (ForecastDay, error) {
		res := sync.Map{}

		wg := sync.WaitGroup{}
		wg.Add(FetchDaysCount)
		for i := range FetchDaysCount {
			go requestFunc(ctx, &wg, lat, lon, &res, i, wp)
		}
		wg.Wait()

		forecast := make(ForecastDay)

		var resErr error

		res.Range(func(key, value any) bool {
			k, ok := key.(string)

			if !ok {
				return false
			}

			switch v := value.(type) {
			case float64:
				forecast[k] = ForecastData{
					Temperature: v,
				}
			case error:
				resErr = v
				return false
			default:
				resErr = fmt.Errorf("unknown result type from the request")
				return false
			}
			return true
		})
		if resErr != nil {
			return nil, resErr
		}
		return forecast, nil
	}
}
