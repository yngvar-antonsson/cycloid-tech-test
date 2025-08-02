package provider

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestOpenMeteoRequest_Success(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	http.DefaultTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return mockHTTPResponse(200, `{
				"daily": {
					"time": ["2025-08-01"],
					"temperature_2m_max": [27.5]
				}
			}`), nil
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	res := &sync.Map{}
	ctx := context.Background()

	openMeteoRequest(ctx, wg, "52.52", "13.41", res, 0, &OpenMeteo{})
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

	if temp != 27.5 {
		t.Errorf("expected temperature 27.5, got %.2f", temp)
	}
}

func TestOpenMeteoRequest_InvalidJSON(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	http.DefaultTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return mockHTTPResponse(200, `{invalid json}`), nil
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	res := &sync.Map{}
	ctx := context.Background()

	openMeteoRequest(ctx, wg, "52.52", "13.41", res, 0, &OpenMeteo{})
	wg.Wait()

	today := time.Now().Format("2006-01-02")
	val, ok := res.Load(today)
	if !ok {
		t.Fatalf("expected error result for %s not found", today)
	}

	if _, ok := val.(error); !ok {
		t.Errorf("expected error, got %T", val)
	}
}

func TestOpenMeteoRequest_ContextCancelled(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	res := &sync.Map{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before call

	openMeteoRequest(ctx, wg, "52.52", "13.41", res, 0, &OpenMeteo{})
	wg.Wait()

	today := time.Now().Format("2006-01-02")
	_, ok := res.Load(today)
	if ok {
		t.Errorf("expected no result due to cancelled context, but got value")
	}
}
