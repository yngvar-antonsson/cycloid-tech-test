package handler

import (
	"context"
	"cycloid/test/provider"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockProvider struct {
	name    string
	data    provider.ForecastDay
	err     error
	timeout time.Duration
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) GetForecast(ctx context.Context, lat, lon string) (provider.ForecastDay, error) {
	if m.timeout > 0 {
		select {
		case <-time.After(m.timeout):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.data, m.err
}

func (m *mockProvider) GetParams(key string) any {
	return nil
}

func (m *mockProvider) SetParams(key string, value any) {}

var originalSettings = settings

func resetSettings() {
	settings = originalSettings
}

func TestAggregateForecast_Success(t *testing.T) {
	p1 := &mockProvider{
		name: "provider1",
		data: provider.ForecastDay{
			"2024-08-01": {Temperature: 25.0},
		},
	}
	p2 := &mockProvider{
		name: "provider2",
		data: provider.ForecastDay{
			"2024-08-01": {Temperature: 27.0},
		},
	}

	providers := []provider.WeatherProvider{p1, p2}
	ctx := context.Background()

	result, err := aggregateForecast(ctx, "52.52", "13.41", providers)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 providers, got %d", len(result))
	}

	if day, ok := result["provider1"]["2024-08-01"]; !ok || day.Temperature != 25.0 {
		t.Errorf("unexpected result from provider1: %+v", day)
	}
}

func TestWeatherHandler_MissingLat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/weather?lon=10", nil)
	w := httptest.NewRecorder()

	WeatherHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
}

func TestWeatherHandler_InvalidLat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/weather?lon=10,lat=300", nil)
	w := httptest.NewRecorder()

	WeatherHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
}

func TestWeatherHandler_MissingLon(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/weather?lat=50", nil)
	w := httptest.NewRecorder()

	WeatherHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
}

func TestWeatherHandler_InvalidLon(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/weather?lon=300,lat=10", nil)
	w := httptest.NewRecorder()

	WeatherHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
}

func TestWeatherHandler_LonIsNotANumber(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/weather?lon=ab,lat=10", nil)
	w := httptest.NewRecorder()

	WeatherHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
}

func TestWeatherHandler_LatIsNotANumber(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/weather?lon=10,lat=ab", nil)
	w := httptest.NewRecorder()

	WeatherHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
}

func TestWeatherHandler_ProviderError(t *testing.T) {
	defer resetSettings()

	settings.Providers = []provider.WeatherProvider{
		&mockProvider{
			name: "badProvider",
			err:  errors.New("fetch error"),
		},
	}
	settings.APILimit = 1

	req := httptest.NewRequest(http.MethodGet, "/weather?lat=50&lon=10", nil)
	w := httptest.NewRecorder()

	WeatherHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got %d", w.Code)
	}
}

func TestWeatherHandler_Success(t *testing.T) {
	defer resetSettings()

	settings.Providers = []provider.WeatherProvider{
		&mockProvider{
			name: "goodProvider",
			data: provider.ForecastDay{
				"2024-08-01": {Temperature: 20.0},
			},
		},
	}
	settings.APILimit = 1

	req := httptest.NewRequest(http.MethodGet, "/weather?lat=50&lon=10", nil)
	w := httptest.NewRecorder()

	WeatherHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var result provider.ProviderForecast
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	forecast, ok := result["goodProvider"]
	if !ok {
		t.Errorf("expected 'goodProvider' key in response")
	}

	day := forecast["2024-08-01"]
	if day.Temperature != 20.0 {
		t.Errorf("unexpected temperature: got %.1f", day.Temperature)
	}
}
