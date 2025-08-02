package handler

import (
	"context"
	"cycloid/test/provider"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func aggregateForecast(ctx context.Context, lat, lon string, providers []provider.WeatherProvider) (provider.ProviderForecast, error) {
	result := make(provider.ProviderForecast)
	for _, p := range providers {
		data, err := p.GetForecast(ctx, lat, lon)
		if err != nil {
			return nil, err
		}
		result[p.Name()] = data
	}
	return result, nil
}

func WeatherHandler(w http.ResponseWriter, r *http.Request) {
	lat := r.URL.Query().Get("lat")
	if lat == "" {
		http.Error(w, "Missing latitude", http.StatusBadRequest)
		return
	}
	if latf, err := strconv.ParseFloat(lat, 32); latf < -90 || latf > 90 || err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
	}
	lon := r.URL.Query().Get("lon")
	if lon == "" {
		http.Error(w, "Missing longitude", http.StatusBadRequest)
		return
	}
	if lonf, err := strconv.ParseFloat(lat, 32); lonf < -180 || lonf > 180 || err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(settings.APILimit)*time.Second)
	defer cancel()

	data, err := aggregateForecast(ctx, lat, lon, settings.Providers)
	if err != nil {
		http.Error(w, "Failed to aggregate forecast: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(data)
}
