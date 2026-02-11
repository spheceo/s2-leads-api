package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type SearchInput struct {
	BusinessType string `json:"business_type" validate:"required,min=2"`
	City         string `json:"city" validate:"required,min=2"`
	CountryCode  string `json:"country_code" validate:"required,min=2"`
	Limit        int64  `json:"limit" validate:"required,gte=1,lte=500"`
}

type CoordinateUpstreamInput []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type cachedCoordinates struct {
	Data      CoordinateUpstreamInput
	ExpiresAt time.Time
}

var (
	coordinatesCacheMu sync.RWMutex
	coordinatesCache   = make(map[string]cachedCoordinates)
	coordinatesTTL     = 6 * time.Hour
)

func GetCoordinates(city, country_code string) (CoordinateUpstreamInput, int, error) {
	cacheKey := strings.ToLower(strings.TrimSpace(city)) + "|" + strings.ToLower(strings.TrimSpace(country_code))
	now := time.Now()

	coordinatesCacheMu.RLock()
	cached, ok := coordinatesCache[cacheKey]
	coordinatesCacheMu.RUnlock()
	if ok && now.Before(cached.ExpiresAt) {
		clone := append(CoordinateUpstreamInput(nil), cached.Data...)
		return clone, http.StatusOK, nil
	}

	httpClient, err := newProxyHTTPClient()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Build the OpenStreetMap endpoint for the requested search input.
	endpoint := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1&countrycodes=%s",
		url.QueryEscape(city), url.QueryEscape(country_code),
	)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to build request: %w", err)
	}

	// Set headers expected by this endpoint.
	req.Header.Set("User-Agent", "LeadGenerator/1.0")
	req.Header.Set("Accepted-Language", "en")

	// Execute upstream request.
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("failed to call nominatim: %w", err)
	}
	defer resp.Body.Close()

	// Read response body once for both validation and decoding.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("failed to read response: %w", err)
	}

	status := resp.StatusCode

	// after reading respBody
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("nominatim returned %d: %s", resp.StatusCode, string(respBody))
	}
	if len(strings.TrimSpace(string(respBody))) == 0 {
		return nil, http.StatusBadGateway, fmt.Errorf("nominatim returned empty body")
	}

	var input CoordinateUpstreamInput
	if err := json.Unmarshal(respBody, &input); err != nil {
		return nil, status, fmt.Errorf("invalid json: %w", err)
	}

	coordinatesCacheMu.Lock()
	coordinatesCache[cacheKey] = cachedCoordinates{
		Data:      append(CoordinateUpstreamInput(nil), input...),
		ExpiresAt: now.Add(coordinatesTTL),
	}
	coordinatesCacheMu.Unlock()

	return input, status, nil
}
