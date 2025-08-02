package provider

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestWeatherAPIRequest_Success(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	http.DefaultTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return mockHTTPResponse(200, `{
				"forecast": {
					"forecastday": [
						{
							"date": "2025-08-01",
							"day": { "maxtemp_c": 29.1 }
						}
					]
				}
			}`), nil
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	res := &sync.Map{}
	ctx := context.Background()

	api := &WeatherAPI{params: map[string]any{"APIKey": "testkey"}}
	weatherAPIRrequest(ctx, wg, "52.52", "13.41", res, 0, api)
	wg.Wait()

	today := time.Now().Format("2006-01-02")
	val, ok := res.Load(today)
	if !ok {
		t.Fatalf("expected result for %s not found", today)
	}

	temp, ok := val.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", val)
	}

	if temp != 29.1 {
		t.Errorf("expected 29.1, got %.2f", temp)
	}
}

func TestWeatherAPIRequest_InvalidJSON(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	http.DefaultTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return mockHTTPResponse(200, `{invalid}`), nil
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	res := &sync.Map{}
	ctx := context.Background()

	api := &WeatherAPI{params: map[string]any{"APIKey": "testkey"}}
	weatherAPIRrequest(ctx, wg, "52.52", "13.41", res, 0, api)
	wg.Wait()

	today := time.Now().Format("2006-01-02")
	val, ok := res.Load(today)
	if !ok {
		t.Fatalf("expected result for %s not found", today)
	}

	if _, ok := val.(error); !ok {
		t.Errorf("expected error, got %T", val)
	}
}

func TestWeatherAPIRequest_ContextCancelled(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	res := &sync.Map{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before a call

	api := &WeatherAPI{params: map[string]any{"APIKey": "testkey"}}
	weatherAPIRrequest(ctx, wg, "52.52", "13.41", res, 0, api)
	wg.Wait()

	today := time.Now().Format("2006-01-02")
	_, ok := res.Load(today)
	if ok {
		t.Errorf("expected no result due to cancelled context, but got value")
	}
}
