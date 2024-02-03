package controllers

import (
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/content"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"github.com/fanky5g/ponzu/internal/util"
	"log"
	"net/http"
	"strings"
)

func NewDeleteHandler(configService config.Service, contentService content.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		appName, err := configService.GetAppName()
		if err != nil {
			log.Printf("Failed to get app name: %v\n", appName)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			LogAndFail(res, err, appName)
			return
		}

		id := req.FormValue("id")
		t := req.FormValue("type")
		ct := t

		if id == "" || t == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		// catch specifier suffix from delete form value
		if strings.Contains(t, "__") {
			spec := strings.Split(t, "__")
			ct = spec[0]
		}

		p, ok := item.Types[ct]
		if !ok {
			log.Println("Type", t, "does not implement item.Hookable or embed item.Item.")
			res.WriteHeader(http.StatusBadRequest)
			errView, err := views.Admin(util.Html("error_400"), appName)
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		post := p()
		hook, ok := post.(item.Hookable)
		if !ok {
			log.Println("Type", t, "does not implement item.Hookable or embed item.Item.")
			res.WriteHeader(http.StatusBadRequest)
			errView, err := views.Admin(util.Html("error_400"), appName)
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		post, err = contentService.GetContent(t, id)
		if err != nil {
			LogAndFail(res, err, appName)
			return
		}

		reject := req.URL.Query().Get("reject")
		if reject == "true" {
			err = hook.BeforeReject(res, req)
			if err != nil {
				log.Println("Error running BeforeReject method in deleteHandler for:", t, err)
				return
			}
		}

		err = hook.BeforeAdminDelete(res, req)
		if err != nil {
			log.Println("Error running BeforeAdminDelete method in deleteHandler for:", t, err)
			return
		}

		err = hook.BeforeDelete(res, req)
		if err != nil {
			log.Println("Error running BeforeDelete method in deleteHandler for:", t, err)
			return
		}

		err = contentService.DeleteContent(t, id)
		if err != nil {
			LogAndFail(res, err, appName)
			return
		}

		err = hook.AfterDelete(res, req)
		if err != nil {
			log.Println("Error running AfterDelete method in deleteHandler for:", t, err)
			return
		}

		err = hook.AfterAdminDelete(res, req)
		if err != nil {
			log.Println("Error running AfterDelete method in deleteHandler for:", t, err)
			return
		}

		if reject == "true" {
			err = hook.AfterReject(res, req)
			if err != nil {
				log.Println("Error running AfterReject method in deleteHandler for:", t, err)
				return
			}
		}

		redir := strings.TrimSuffix(req.URL.Scheme+req.URL.Host+req.URL.Path, "/edit/delete")
		redir = redir + "/contents?type=" + ct
		http.Redirect(res, req, redir, http.StatusFound)
	}
}
