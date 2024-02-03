package controllers

import (
	"errors"
	"fmt"
	"github.com/fanky5g/ponzu/internal/application/auth"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/users"
	domainErrors "github.com/fanky5g/ponzu/internal/domain/errors"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"github.com/fanky5g/ponzu/internal/util"
	emailer "github.com/nilslice/email"
	"log"
	"net/http"
	"strings"
)

func NewForgotPasswordHandler(
	configService config.Service,
	userService users.Service,
	authService auth.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		appName, err := configService.GetAppName()
		if err != nil {
			log.Printf("Failed to get app name: %v\n", appName)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch req.Method {
		case http.MethodGet:
			view, err := views.ForgotPassword(appName)
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			res.Write(view)

		case http.MethodPost:
			err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			// check email for user, if no user return Error
			email := strings.ToLower(req.FormValue("email"))
			if email == "" {
				res.WriteHeader(http.StatusBadRequest)
				log.Println("Failed account recovery. No email address submitted.")
				return
			}

			_, err = userService.GetUserByEmail(email)
			if errors.Is(err, domainErrors.ErrNoUserExists) {
				res.WriteHeader(http.StatusBadRequest)
				log.Println("No user exists.", err)
				return
			}

			// create temporary key to verify user
			key, err := authService.SetRecoveryKey(email)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				log.Println("Failed to set account recovery key.", err)
				return
			}

			domain, err := configService.GetStringValue("domain")
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				log.Println("Failed to get domain from configuration.", err)
				return
			}

			body := fmt.Sprintf(`
There has been an account recovery request made for the user with email:
%s

To recover your account, please go to http://%s/recover/key and enter 
this email address along with the following secret key:

%s

If you did not make the request, ignore this message and your password 
will remain as-is.


Thank you,
Ponzu CMS at %s

`, email, domain, key, domain)

			msg := emailer.Message{
				To:      email,
				From:    fmt.Sprintf("ponzu@%s", domain),
				Subject: fmt.Sprintf("Account Recovery [%s]", domain),
				Body:    body,
			}

			go func() {
				err = msg.Send()
				if err != nil {
					log.Println("Failed to send message to:", msg.To, "about", msg.Subject, "Error:", err)
				}
			}()

			// redirect to /recover/key and send email with key and URL
			http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/recover/key", http.StatusFound)

		default:
			res.WriteHeader(http.StatusMethodNotAllowed)
			errView, err := views.Admin(util.Html("error_405"), appName)
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}
	}
}
