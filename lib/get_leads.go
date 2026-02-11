package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
)

type LeadsUpstreamInput struct {
	Data []struct {
		Name        string   `json:"name"`
		PhoneNumber string   `json:"phone_number"`
		FullAddress string   `json:"full_address"`
		City        string   `json:"city"`
		Rating      float64  `json:"rating"`
		ReviewCount int64    `json:"review_count"`
		Website     string   `json:"website"`
		PlaceLink   string   `json:"place_link"`
	} `json:"data"`
}

type LeadOutputItem struct {
	Name        string   `json:"name"`
	PhoneNumber string   `json:"phone_number"`
	Address     string   `json:"address"`
	City        string   `json:"city"`
	Rating      float64  `json:"rating"`
	Reviews     int64    `json:"reviews"`
	Website     string   `json:"website"`
	Link        string   `json:"link"`
	LeadScore   float64  `json:"lead_score"`
}

type LeadsOutput struct {
	Data []LeadOutputItem `json:"data"`
}

func GetLeads(lat, lon, business_type, country_code string, limit int64) (LeadsOutput, int, error) {
	// Get RapidAPI Key
	rapidapi, err := GetEnv("RAPIDAPI_KEY")
	if err != nil {
		return LeadsOutput{}, http.StatusInternalServerError, fmt.Errorf("ENV failed: %w", err)
	}

	httpClient, err := newProxyHTTPClient()
	if err != nil {
		return LeadsOutput{}, http.StatusInternalServerError, err
	}

	endpoint := fmt.Sprintf(
		"https://maps-data.p.rapidapi.com/searchmaps.php?query=%s&limit=%d&country=%s&lang=en&offset=0&zoom=13&lat=%s&lng=%s",
		url.QueryEscape(business_type), limit, country_code, lat, lon,
	)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return LeadsOutput{}, http.StatusInternalServerError, fmt.Errorf("failed to build request: %w", err)
	}

	// Set headers expected by this endpoint.
	req.Header.Set("x-rapidapi-key", rapidapi)

	// Execute upstream request.
	resp, err := httpClient.Do(req)
	if err != nil {
		return LeadsOutput{}, http.StatusBadGateway, fmt.Errorf("failed to call rapidapi: %w", err)
	}
	defer resp.Body.Close()

	// Read response body once for both validation and decoding.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return LeadsOutput{}, http.StatusBadGateway, fmt.Errorf("failed to read response: %w", err)
	}

	status := resp.StatusCode

	// after reading respBody
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return LeadsOutput{}, resp.StatusCode, fmt.Errorf("rapidapi returned %d: %s", resp.StatusCode, string(respBody))
	}
	if len(strings.TrimSpace(string(respBody))) == 0 {
		return LeadsOutput{}, http.StatusBadGateway, fmt.Errorf("rapidapi returned empty body")
	}

	var input LeadsUpstreamInput
	if err := json.Unmarshal(respBody, &input); err != nil {
		return LeadsOutput{}, status, fmt.Errorf("invalid json: %w", err)
	}

	body := LeadsOutput{
		Data: make([]LeadOutputItem, 0, len(input.Data)),
	}

	for _, item := range input.Data {
		body.Data = append(body.Data, LeadOutputItem{
			Name:        item.Name,
			PhoneNumber: item.PhoneNumber,
			Address:     item.FullAddress,
			City:        item.City,
			Rating:      item.Rating,
			Reviews:     item.ReviewCount,
			Website:     item.Website,
			Link:        item.PlaceLink,
			LeadScore:   calculateLeadScore(item.Rating, item.ReviewCount),
		})
	}

	return body, status, nil
}

func calculateLeadScore(rating float64, reviews int64) float64 {
	clampedRating := math.Max(0, math.Min(5, rating))
	badRating := (5 - clampedRating) / 5
	lowReviewFactor := 1 / math.Log10(float64(reviews)+10)

	// Weighted additive model:
	// - rating contributes 35%
	// - low review count contributes 65%
	// This keeps high scores possible even with perfect rating if review count is very low.
	score := 100 * ((0.35 * badRating) + (0.65 * lowReviewFactor))

	// Keep API output readable and stable.
	return math.Round(score*100) / 100
}
