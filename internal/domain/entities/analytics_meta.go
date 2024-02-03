package entities

type AnalyticsMetric struct {
	Date   string `json:"date"`
	Total  int    `json:"total"`
	Unique int    `json:"unique"`
}

type AnalyticsHTTPRequestMetadata struct {
	URL        string `json:"url"`
	Method     string `json:"http_method"`
	Origin     string `json:"origin"`
	Proto      string `json:"http_protocol"`
	RemoteAddr string `json:"ip_address"`
	Timestamp  int64  `json:"timestamp"`
	External   bool   `json:"external_content"`
}
