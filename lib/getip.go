package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func newProxyHTTPClient() (*http.Client, error) {
	proxyURL, err := GetEnv("PROXY_URL")
	if err != nil {
		return nil, fmt.Errorf("ENV failed: %w", err)
	}
	if rest, ok := strings.CutPrefix(proxyURL, "https://"); ok {
		proxyURL = "http://" + rest
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}
	if proxyURL != "" {
		parsedProxyURL, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("proxy parse error: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsedProxyURL)
	}

	httpClient := &http.Client{
		Timeout:   25 * time.Second,
		Transport: transport,
	}

	return httpClient, nil
}

func GetIP() (any, int, error) {
	httpClient, err := newProxyHTTPClient()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	ipReq, err := http.NewRequest(http.MethodGet, "https://api.ipify.org?format=json", nil)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create ip request: %w", err)
	}
	ipReq.Header.Set("accept", "application/json")

	ipResp, err := httpClient.Do(ipReq)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("ip request failed: %w", err)
	}
	defer ipResp.Body.Close()

	ipBody, err := io.ReadAll(ipResp.Body)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("failed reading ip response: %w", err)
	}
	if ipResp.StatusCode < 200 || ipResp.StatusCode >= 300 {
		return nil, ipResp.StatusCode, fmt.Errorf("ip probe returned status %d", ipResp.StatusCode)
	}

	var ipData ipifyResponse
	if err := json.Unmarshal(ipBody, &ipData); err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("failed decoding ip response: %w", err)
	}
	if ipData.IP == "" {
		return nil, http.StatusBadGateway, fmt.Errorf("ip probe returned empty ip")
	}

	geoReq, err := http.NewRequest(http.MethodGet, "https://ipwho.is/"+url.QueryEscape(ipData.IP), nil)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create geo request: %w", err)
	}
	geoReq.Header.Set("accept", "application/json")

	geoResp, err := httpClient.Do(geoReq)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("geo request failed: %w", err)
	}
	defer geoResp.Body.Close()

	geoBody, err := io.ReadAll(geoResp.Body)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("failed reading geo response: %w", err)
	}
	if geoResp.StatusCode < 200 || geoResp.StatusCode >= 300 {
		return nil, geoResp.StatusCode, fmt.Errorf("geo probe returned status %d", geoResp.StatusCode)
	}

	var geo ipWhoIsResponse
	if err := json.Unmarshal(geoBody, &geo); err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("failed decoding geo response: %w", err)
	}
	if !geo.Success {
		return nil, http.StatusBadGateway, fmt.Errorf("geo probe failed: %s", geo.Message)
	}

	body := ProxyIdentity{
		IP:      ipData.IP,
		City:    geo.City,
		Region:  geo.Region,
		Country: geo.Country,
		Org:     geo.Connection.Org,
	}
	return body, http.StatusOK, nil
}
