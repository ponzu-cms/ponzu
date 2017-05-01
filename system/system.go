// Package system contains a collection of packages that make up the internal
// Ponzu system, which handles addons, administration, the Admin server, the API
// server, analytics, databases, search, TLS, and various internal types.
package system

import (
	"net/http"

	"github.com/ponzu-cms/ponzu/system/db"
)

// BasicAuth adds HTTP Basic Auth check for requests that should implement it
func BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		u := db.ConfigCache("backup_basic_auth_user").(string)
		p := db.ConfigCache("backup_basic_auth_password").(string)

		if u == "" || p == "" {
			res.WriteHeader(http.StatusForbidden)
			return
		}

		user, password, ok := req.BasicAuth()

		if !ok {
			res.WriteHeader(http.StatusForbidden)
			return
		}

		if u != user || p != password {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(res, req)
	})
}
