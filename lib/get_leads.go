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
		"https://maps-data.p.rapidapi.com/searchmaps.php?query=%s&limit=%d&country=%s&lang=en&offset=0&zoom=10&lat=%s&lng=%s",
		url.QueryEscape(business_type), limit, country_code, lat, lon,
	)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return LeadsOutput{}, http.StatusInternalServerError, fmt.Errorf("failed to build request: %w", err)
	}

	// Set headers expected by this endpoint.
	req.Header.Set("x-rapidapi-key", rapidapi)
	req.Header.Set("x-rapidapi-host", "maps-data.p.rapidapi.com")

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
		Total: len(input.Data),
		Data:  make([]LeadOutputItem, 0, len(input.Data)),
	}

	for _, item := range input.Data {
		body.Data = append(body.Data, LeadOutputItem{
			BusinessID:  item.BusinessID,
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

func GetReviews(businessID string, limit int64) (ReviewsOutput, int, error) {
	rapidapi, err := GetEnv("RAPIDAPI_KEY")
	if err != nil {
		return ReviewsOutput{}, http.StatusInternalServerError, fmt.Errorf("ENV failed: %w", err)
	}

	httpClient, err := newProxyHTTPClient()
	if err != nil {
		return ReviewsOutput{}, http.StatusInternalServerError, err
	}

	body, status, err := fetchReviews(httpClient, rapidapi, businessID, limit)
	if err != nil {
		return ReviewsOutput{}, status, err
	}
	if body.Total > 0 {
		return body, status, nil
	}

	return fetchReviews(httpClient, rapidapi, businessID, limit)
}

func fetchReviews(httpClient *http.Client, rapidapi, businessID string, limit int64) (ReviewsOutput, int, error) {
	endpoint := fmt.Sprintf(
		"https://maps-data.p.rapidapi.com/reviews.php?business_id=%s&lang=en&limit=%d&sort=Newest",
		url.QueryEscape(businessID), limit,
	)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return ReviewsOutput{}, http.StatusInternalServerError, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("x-rapidapi-key", rapidapi)
	req.Header.Set("x-rapidapi-host", "maps-data.p.rapidapi.com")

	resp, err := httpClient.Do(req)
	if err != nil {
		return ReviewsOutput{}, http.StatusBadGateway, fmt.Errorf("failed to call rapidapi reviews: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReviewsOutput{}, http.StatusBadGateway, fmt.Errorf("failed to read reviews response: %w", err)
	}

	status := resp.StatusCode

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ReviewsOutput{}, resp.StatusCode, fmt.Errorf("rapidapi reviews returned %d: %s", resp.StatusCode, string(respBody))
	}
	if len(strings.TrimSpace(string(respBody))) == 0 {
		return ReviewsOutput{}, http.StatusBadGateway, fmt.Errorf("rapidapi reviews returned empty body")
	}

	var input ReviewsUpstreamInput
	if err := json.Unmarshal(respBody, &input); err != nil {
		return ReviewsOutput{}, status, fmt.Errorf("invalid reviews json: %w", err)
	}

	upstreamReviews := input.Data.Reviews
	reviews := make([]ReviewItem, 0, len(upstreamReviews))
	for _, item := range upstreamReviews {
		text := strings.TrimSpace(item.ReviewText)
		if text == "" {
			continue
		}

		date := item.ISODate
		if date == "" {
			date = item.ReviewTime
		}

		reviews = append(reviews, ReviewItem{
			Author: item.UserName,
			Rating: item.ReviewRate,
			Text:   text,
			Date:   date,
		})
	}

	reviewsOutput := ReviewsOutput{
		Total: len(reviews),
		Data:  reviews,
	}

	return reviewsOutput, status, nil
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
