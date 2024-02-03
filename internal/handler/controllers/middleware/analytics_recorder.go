package middleware

import (
	"github.com/fanky5g/ponzu/internal/application/analytics"
	"github.com/fanky5g/ponzu/internal/handler/controllers/mappers"
	"net/http"
)

var AnalyticsRecorderMiddleware Token = "AnalyticsRecorderMiddleware"

func NewAnalyticsRecorderMiddleware(analyticsService analytics.Service) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			request := mappers.GetAnalyticsRequestMetadata(req)
			go analyticsService.Record(request)

			next.ServeHTTP(res, req)
		}
	}
}
