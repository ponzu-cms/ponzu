package controllers

import (
	"github.com/fanky5g/ponzu/internal/application/auth"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/users"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"log"
	"net/http"
	"strings"
)

func NewRecoveryKeyHandler(
	configService config.Service,
	authService auth.Service,
	userService users.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			appName, err := configService.GetAppName()
			if err != nil {
				log.Printf("Failed to get app name: %v\n", appName)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			view, err := views.RecoveryKey(appName)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			res.Write(view)

		case http.MethodPost:
			err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
			if err != nil {
				log.Println("Error parsing recovery key form:", err)

				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte("Error, please go back and try again."))
				return
			}

			// check for email & key match
			email := strings.ToLower(req.FormValue("email"))
			key := req.FormValue("key")

			var actual string
			if actual, err = authService.GetRecoveryKey(email); err != nil || actual == "" {
				log.Println("Error getting recovery key from database:", err)

				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte("Error, please go back and try again."))
				return
			}

			if key != actual {
				log.Println("Bad recovery key submitted:", key)

				res.WriteHeader(http.StatusBadRequest)
				res.Write([]byte("Error, please go back and try again."))
				return
			}

			// set user with new password
			password := req.FormValue("password")
			var user *entities.User
			user, err = userService.GetUserByEmail(email)
			if err != nil {
				log.Println("Error finding user by email:", email, err)

				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte("Error, please go back and try again."))
				return
			}

			if user == nil {
				log.Println("No user found with email:", email)

				res.WriteHeader(http.StatusBadRequest)
				res.Write([]byte("Error, please go back and try again."))
				return
			}

			if err = authService.SetCredential(user.ID, &entities.Credential{
				Type:  entities.CredentialTypePassword,
				Value: password,
			}); err != nil {
				log.Println("Error updating user:", err)

				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte("Error, please go back and try again."))
				return
			}

			// redirect to /login
			redir := req.URL.Scheme + req.URL.Host + "/login"
			http.Redirect(res, req, redir, http.StatusFound)

		default:
			res.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}
