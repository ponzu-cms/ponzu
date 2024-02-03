package controllers

import (
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/storage"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"github.com/fanky5g/ponzu/internal/util"
	"log"
	"net/http"
)

func NewDeleteUploadHandler(configService config.Service, storageService storage.Service) http.HandlerFunc {
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
		if id == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		post := interface{}(&item.FileUpload{})
		hook, ok := post.(item.Hookable)
		if !ok {
			log.Println("Type", storage.UploadsEntityName, "does not implement item.Hookable or embed item.Item.")
			res.WriteHeader(http.StatusBadRequest)
			errView, err := views.Admin(util.Html("error_400"), appName)
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		err = hook.BeforeDelete(res, req)
		if err != nil {
			log.Println("Error running BeforeDelete method in deleteHandler for:", storage.UploadsEntityName, err)
			return
		}

		err = storageService.DeleteFile(id)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = hook.AfterDelete(res, req)
		if err != nil {
			log.Println("Error running AfterDelete method in deleteHandler for:", storage.UploadsEntityName, err)
			return
		}

		redir := "/uploads"
		http.Redirect(res, req, redir, http.StatusFound)
	}
}
