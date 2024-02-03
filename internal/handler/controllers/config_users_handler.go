package controllers

import (
	"bytes"
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

// UsersList ...
func UsersList(
	userService users.Service,
	currentUser *entities.User,
) ([]byte, error) {
	systemUsers, err := userService.ListUsers()
	if err != nil {
		return nil, err
	}

	for i, user := range systemUsers {
		if user.Email == currentUser.Email {
			systemUsers = append(systemUsers[:i], systemUsers[i+1:]...)
		}
	}

	// make buffer to execute views into then pass buffer's bytes to Admin
	buf := &bytes.Buffer{}
	tmpl := util.MakeTemplate("users_list")
	data := map[string]interface{}{
		"GetUserByEmail": currentUser,
		"Users":          systemUsers,
	}

	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}

	return views.Admin(buf.String(), "")
}

func NewConfigUsersHandler(
	configService config.Service,
	authService auth.Service,
	userService users.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// TODO: this is called a lot in many places. We can put this into a middleware
		// 	that sets appName on the request context
		// 	as a global middleware: https://www.jvt.me/posts/2023/09/01/golang-nethttp-global-middleware/
		appName, err := configService.GetAppName()
		if err != nil {
			log.Printf("Failed to get app name: %v\n", appName)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch req.Method {
		case http.MethodGet:
			currentUser, err := authService.GetUserFromAuthToken(mappers.GetAuthToken(req))
			if err != nil {
				LogAndFail(res, fmt.Errorf("failed to get current user: %v", err), appName)
				return
			}

			view, err := UsersList(userService, currentUser)
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			res.Write(view)

		case http.MethodPost:
			// create new user
			err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			email := strings.ToLower(req.FormValue("email"))
			password := req.PostFormValue("password")
			if email == "" || password == "" {
				LogAndFail(res, err, appName)
				return
			}

			// TODO: UnitOfWork. We need to be able to run all or fail all operations
			user, err := userService.CreateUser(email)
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			if err = authService.SetCredential(user.ID, &entities.Credential{
				Type:  entities.CredentialTypePassword,
				Value: password,
			}); err != nil {
				LogAndFail(res, err, appName)
				return
			}

			http.Redirect(res, req, req.URL.String(), http.StatusFound)

		default:
			res.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
