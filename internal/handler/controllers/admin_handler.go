package controllers

import (
	"bytes"
	"github.com/fanky5g/ponzu/internal/application/analytics"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"github.com/fanky5g/ponzu/internal/util"
	"log"
	"net/http"
)

// Dashboard returns the controllers view with analytics dashboard
func Dashboard(analyticsService analytics.Service, configService config.Service) ([]byte, error) {
	buf := &bytes.Buffer{}
	data, err := analyticsService.GetChartData()
	if err != nil {
		return nil, err
	}

	appName, err := configService.GetAppName()
	if err != nil {
		return nil, err
	}

	tmpl := util.MakeTemplate("analytics")
	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}

	return views.Admin(buf.String(), appName)
}

func NewAdminHandler(analyticsService analytics.Service, configService config.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		view, err := Dashboard(analyticsService, configService)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(view)
	}
}
