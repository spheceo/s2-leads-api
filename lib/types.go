package lib

// CoordinateUpstreamInput models the Nominatim geocoding response.
type CoordinateUpstreamInput []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// LeadsUpstreamInput models the RapidAPI maps search response.
type LeadsUpstreamInput struct {
	Data []struct {
		Name        string  `json:"name"`
		PhoneNumber string  `json:"phone_number"`
		FullAddress string  `json:"full_address"`
		City        string  `json:"city"`
		Rating      float64 `json:"rating"`
		ReviewCount int64   `json:"review_count"`
		Website     string  `json:"website"`
		PlaceLink   string  `json:"place_link"`
	} `json:"data"`
}

// LeadOutputItem is the normalized lead returned to API clients.
type LeadOutputItem struct {
	Name        string  `json:"name"`
	PhoneNumber string  `json:"phone_number"`
	Address     string  `json:"address"`
	City        string  `json:"city"`
	Rating      float64 `json:"rating"`
	Reviews     int64   `json:"reviews"`
	Website     string  `json:"website"`
	Link        string  `json:"link"`
	LeadScore   float64 `json:"lead_score"`
}

// LeadsOutput is the final response shape for /search.
type LeadsOutput struct {
	Total int              `json:"total"`
	Data  []LeadOutputItem `json:"data"`
}

// ProxyIdentity is the combined response from IP + geo lookup.
type ProxyIdentity struct {
	IP      string `json:"ip"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

// ipifyResponse models the response from api.ipify.org.
type ipifyResponse struct {
	IP string `json:"ip"`
}

// ipWhoIsResponse models the response from ipwho.is.
type ipWhoIsResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	IP         string `json:"ip"`
	City       string `json:"city"`
	Region     string `json:"region"`
	Country    string `json:"country"`
	Connection struct {
		Org string `json:"org"`
	} `json:"connection"`
}
