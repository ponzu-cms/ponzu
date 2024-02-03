package controllers

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/application/auth"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/users"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"github.com/fanky5g/ponzu/internal/handler/controllers/mappers"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"github.com/fanky5g/ponzu/internal/util"
	"log"
	"net/http"
	"strings"
)

func NewConfigUsersEditHandler(
	configService config.Service,
	authService auth.Service,
	userService users.Service,
) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			appName, err := configService.GetAppName()
			if err != nil {
				log.Printf("Failed to get app name: %v\n", appName)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := views.Admin(util.Html("error_500"), appName)
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			user, err := authService.GetUserFromAuthToken(mappers.GetAuthToken(req))
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			// check if password matches
			password := &entities.Credential{
				Type:  entities.CredentialTypePassword,
				Value: req.PostFormValue("password"),
			}

			if err = authService.VerifyCredential(user.ID, password); err != nil {
				log.Printf("Unexpected user/password combination: %v\n", err)
				res.WriteHeader(http.StatusBadRequest)
				errView, err := views.Admin(util.Html("error_405"), appName)
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			email := strings.ToLower(req.PostFormValue("email"))
			newPassword := req.PostFormValue("new_password")
			if newPassword != "" {
				if err = authService.SetCredential(user.ID, &entities.Credential{
					Type:  entities.CredentialTypePassword,
					Value: newPassword,
				}); err != nil {
					LogAndFail(res, fmt.Errorf("failed to update password: %v", err), appName)
					return
				}
			}

			if email != "" {
				update := &entities.User{
					ID:    user.ID,
					Email: email,
				}

				if err = userService.UpdateUser(user, update); err != nil {
					LogAndFail(res, fmt.Errorf("failed to update user: %v", err), appName)
					return
				}

				user = update
			}

			// create new token
			token, expires, err := authService.NewToken(user)
			if err != nil {
				LogAndFail(res, fmt.Errorf("failed to generate token: %v", err), appName)
				return
			}

			cookie := &http.Cookie{
				Name:    "_token",
				Value:   token,
				Expires: expires,
				Path:    "/",
			}

			http.SetCookie(res, cookie)
			// add new token cookie to the request
			req.AddCookie(cookie)
			http.Redirect(res, req, strings.TrimSuffix(req.URL.String(), "/edit"), http.StatusFound)

		default:
			res.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
