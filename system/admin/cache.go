package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bosssauce/ponzu/system/db"
)

// CacheControl sets the default cache policy on static asset responses
func CacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		etag := db.ConfigCache("etag")
		policy := fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 60*60*24*30)
		res.Header().Add("Etag", etag)
		res.Header().Add("Cache-Control", policy)

		if match := res.Header().Get("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				res.WriteHeader(http.StatusNotModified)
				return
			}
		}

		next.ServeHTTP(res, req)
	})
}
