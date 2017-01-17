package db

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CacheControl sets the default cache policy on static asset responses
func CacheControl(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		etag := ConfigCache("etag").(string)
		policy := fmt.Sprintf("max-age=%d, public", 60*60*24*30)
		res.Header().Add("ETag", etag)
		res.Header().Add("Cache-Control", policy)

		if match := req.Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				res.WriteHeader(http.StatusNotModified)
				return
			}
		}

		next.ServeHTTP(res, req)
	})
}

// NewEtag generates a new Etag for response caching
func NewEtag() string {
	now := fmt.Sprintf("%d", time.Now().Unix())
	etag := base64.StdEncoding.EncodeToString([]byte(now))

	return etag
}

// InvalidateCache sets a new Etag for http responses
func InvalidateCache() error {
	err := PutConfig("etag", NewEtag())
	if err != nil {
		return err
	}

	return nil
}
