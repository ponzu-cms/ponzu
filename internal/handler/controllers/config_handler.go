package controllers

import (
	"encoding/json"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"log"
	"net/http"
)

func NewConfigHandler(configService config.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			data, err := configService.GetAll()
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			c := &entities.Config{}

			err = json.Unmarshal(data, c)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			cfg, err := c.MarshalEditor()
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			appName, err := configService.GetAppName()
			if err != nil {
				log.Printf("Failed to get app name: %v\n", appName)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			adminView, err := views.Admin(string(cfg), appName)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			res.Header().Set("Content-Type", "text/html")
			res.Write(adminView)

		case http.MethodPost:
			err := req.ParseForm()
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = configService.SetConfig(req.Form)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.Redirect(res, req, req.URL.String(), http.StatusFound)

		default:
			res.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
