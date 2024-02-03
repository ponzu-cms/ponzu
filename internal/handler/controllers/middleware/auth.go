package middleware

import (
	"github.com/fanky5g/ponzu/internal/application/auth"
	"github.com/fanky5g/ponzu/internal/handler/controllers/mappers"
	"log"
	"net/http"
)

var AuthMiddleware Token = "AuthMiddleware"

func NewAuthMiddleware(authService auth.Service) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			authToken := mappers.GetAuthToken(req)
			isValid, err := authService.IsTokenValid(authToken)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				log.Printf("Failed to check token validity: %v\n", err)
				return
			}

			if isValid {
				next.ServeHTTP(res, req)
				return
			}

			redir := req.URL.Scheme + req.URL.Host + "/login"
			http.Redirect(res, req, redir, http.StatusFound)
		}
	}
}
