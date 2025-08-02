package provider

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

type mockProvider struct{}

func (m *mockProvider) GetParams(key string) any        { return nil }
func (m *mockProvider) SetParams(key string, value any) {}
func (m *mockProvider) Name() string                    { return "mock" }
func (m *mockProvider) GetForecast(ctx context.Context, lat, lon string) (ForecastDay, error) {
	return nil, nil
}

func TestBuildGetter_Success(t *testing.T) {
	mockReqFunc := func(ctx context.Context, wg *sync.WaitGroup, lat, lon string, res *sync.Map, i int, wp WeatherProvider) {
		defer wg.Done()
		// simulate delay
		time.Sleep(10 * time.Millisecond)
		res.Store(fmt.Sprintf("day%d", i), float64(i+10)) // predictable data
	}

	getter := BuildForecastGetter(context.Background(), "52.52", "13.41", &mockProvider{}, mockReqFunc)

	result, err := getter(context.Background(), "52.52", "13.41")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(result) != FetchDaysCount {
		t.Fatalf("expected %d days, got %d", FetchDaysCount, len(result))
	}

	for i := 0; i < FetchDaysCount; i++ {
		day := fmt.Sprintf("day%d", i)
		data, ok := result[day]
		if !ok {
			t.Errorf("missing day: %s", day)
			continue
		}
		expectedTemp := float64(i + 10)
		if data.Temperature != expectedTemp {
			t.Errorf("expected temp %.2f for %s, got %.2f", expectedTemp, day, data.Temperature)
		}
	}
}

func TestBuildGetter_ErrorFromRequest(t *testing.T) {
	mockErr := errors.New("mock error")
	mockReqFunc := func(ctx context.Context, wg *sync.WaitGroup, lat, lon string, res *sync.Map, i int, wp WeatherProvider) {
		defer wg.Done()
		if i == 2 {
			res.Store("day2", mockErr)
		} else {
			res.Store(fmt.Sprintf("day%d", i), float64(i))
		}
	}

	getter := BuildForecastGetter(context.Background(), "44.0", "10.0", &mockProvider{}, mockReqFunc)

	_, err := getter(context.Background(), "44.0", "10.0")
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	if err.Error() != "mock error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildGetter_InvalidResultType(t *testing.T) {
	mockReqFunc := func(ctx context.Context, wg *sync.WaitGroup, lat, lon string, res *sync.Map, i int, wp WeatherProvider) {
		defer wg.Done()
		res.Store("day0", struct{}{}) // invalid type
	}

	getter := BuildForecastGetter(context.Background(), "44.0", "10.0", &mockProvider{}, mockReqFunc)

	_, err := getter(context.Background(), "44.0", "10.0")
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	if err.Error() != "unknown result type from the request" {
		t.Errorf("unexpected error: %v", err)
	}
}
