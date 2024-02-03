package controllers

import (
	"github.com/fanky5g/ponzu/internal/application/auth"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/users"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"github.com/fanky5g/ponzu/internal/handler/controllers/mappers"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"log"
	"net/http"
	"strings"
)

func hasSystemUsers(userService users.Service) (bool, error) {
	systemUsers, err := userService.ListUsers()
	if err != nil {
		return false, err
	}

	return len(systemUsers) > 0, nil
}

func NewLoginHandler(
	configService config.Service,
	authService auth.Service,
	userService users.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		systemInitialized, err := hasSystemUsers(userService)
		if err != nil {
			log.Printf("Failed to check system initialization: %v\n", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !systemInitialized {
			redir := req.URL.Scheme + req.URL.Host + "/init"
			http.Redirect(res, req, redir, http.StatusFound)
			return
		}

		isValid, err := authService.IsTokenValid(mappers.GetAuthToken(req))
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Printf("Failed to check token validity: %v\n", err)
			return
		}

		if isValid {
			http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
			return
		}

		switch req.Method {
		case http.MethodGet:
			appName, err := configService.GetAppName()
			if err != nil {
				log.Printf("Failed to get app name: %v\n", appName)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			view, err := views.Login(appName)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			res.Header().Set("Content-Type", "text/html")
			res.Write(view)

		case http.MethodPost:
			err = req.ParseForm()
			if err != nil {
				log.Println(err)
				http.Redirect(res, req, req.URL.String(), http.StatusFound)
				return
			}

			email := strings.ToLower(req.FormValue("email"))
			password := req.FormValue("password")
			authToken, tokenExpiry, err := authService.LoginByEmail(email, &entities.Credential{
				Type:  entities.CredentialTypePassword,
				Value: password,
			})

			if err != nil || authToken == "" {
				log.Println("Failed to login user", err)
				http.Redirect(res, req, req.URL.String(), http.StatusFound)
				return
			}

			http.SetCookie(res, &http.Cookie{
				Name:    "_token",
				Value:   authToken,
				Expires: tokenExpiry,
				Path:    "/",
			})

			http.Redirect(res, req, strings.TrimSuffix(req.URL.String(), "/login"), http.StatusFound)
		}
	}
}
