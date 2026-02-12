package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func GetCoordinates(city, country_code string) (CoordinateUpstreamInput, int, error) {
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

	return input, status, nil
}
