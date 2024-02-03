package controllers

import (
	"encoding/base64"
	"github.com/fanky5g/ponzu/internal/application/auth"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/users"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"github.com/fanky5g/ponzu/internal/util"
	"log"
	"net/http"
	"strings"
)

func NewInitHandler(configService config.Service, userService users.Service, authService auth.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		systemInitialized, err := hasSystemUsers(userService)
		if err != nil {
			log.Printf("Failed to check system initialization: %v\n", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if systemInitialized {
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

			view, err := views.Init(appName)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "text/html")
			res.Write(view)

		case http.MethodPost:
			err := req.ParseForm()
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			// get the site name from post to encode and use as secret
			name := []byte(req.FormValue("name") + util.NewEtag())
			secret := base64.StdEncoding.EncodeToString(name)
			req.Form.Set("client_secret", secret)

			// generate an Etag to use for response caching
			etag := util.NewEtag()
			req.Form.Set("etag", etag)

			// create and save controllers user
			email := strings.ToLower(req.FormValue("email"))
			password := req.FormValue("password")
			user, err := userService.CreateUser(email)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			if err = authService.SetCredential(user.ID, &entities.Credential{
				Type:  entities.CredentialTypePassword,
				Value: password,
			}); err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			// set HTTP port which should be previously added to config cache
			var port string
			port, err = configService.GetCacheStringValue("http_port")
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			req.Form.Set("http_port", port)
			// set initial user email as admin_email and make config
			req.Form.Set("admin_email", email)
			err = configService.SetConfig(req.Form)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			token, expires, err := authService.NewToken(user)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.SetCookie(res, &http.Cookie{
				Name:    "_token",
				Value:   token,
				Expires: expires,
				Path:    "/",
			})

			redir := strings.TrimSuffix(req.URL.String(), "/init")
			http.Redirect(res, req, redir, http.StatusFound)

		default:
			res.WriteHeader(http.StatusMethodNotAllowed)
		}
	}

}
