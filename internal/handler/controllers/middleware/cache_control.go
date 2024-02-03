package middleware

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"net/http"
	"strings"
)

const (
	// DefaultMaxAge provides a 2592000 second (30-day) cache max-age setting
	DefaultMaxAge = int64(60 * 60 * 24 * 30)
)

var CacheControlMiddleware Token = "CacheControlMiddleware"

func NewCacheControlMiddleware(cacheable interfaces.Cacheable) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		cache := cacheable.Cache()

		return func(res http.ResponseWriter, req *http.Request) {
			cacheDisabled := cache.GetByKey("cache_disabled").(bool)
			if cacheDisabled {
				res.Header().Add("Cache-Control", "no-cache")
				next.ServeHTTP(res, req)
			} else {
				age := int64(cache.GetByKey("cache_max_age").(float64))
				etag := cache.GetByKey("etag").(string)
				if age == 0 {
					age = DefaultMaxAge
				}

				policy := fmt.Sprintf("max-age=%d, public", age)
				res.Header().Add("ETag", etag)
				res.Header().Add("Cache-Control", policy)

				if match := req.Header.Get("If-None-Match"); match != "" {
					if strings.Contains(match, etag) {
						res.WriteHeader(http.StatusNotModified)
						return
					}
				}

				next.ServeHTTP(res, req)
			}
		}
	}
}

func ToHttpHandler(middleware Middleware) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return middleware(func(res http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(res, req)
		})
	}
}
