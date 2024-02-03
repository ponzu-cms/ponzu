package mappers

import (
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"net/http"
	"strings"
	"time"
)

func GetAnalyticsRequestMetadata(req *http.Request) entities.AnalyticsHTTPRequestMetadata {
	external := strings.Contains(req.URL.Path, "/external/")
	ts := int64(time.Nanosecond) * time.Now().UnixNano() / int64(time.Millisecond)

	return entities.AnalyticsHTTPRequestMetadata{
		URL:        req.URL.String(),
		Method:     req.Method,
		Origin:     req.Header.Get("Origin"),
		Proto:      req.Proto,
		RemoteAddr: req.RemoteAddr,
		Timestamp:  ts,
		External:   external,
	}
}
