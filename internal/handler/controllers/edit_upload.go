package controllers

import (
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/storage"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"github.com/fanky5g/ponzu/internal/domain/services/management/editor"
	"github.com/fanky5g/ponzu/internal/domain/services/management/manager"
	"github.com/fanky5g/ponzu/internal/handler/controllers/mappers"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"github.com/fanky5g/ponzu/internal/util"
	"log"
	"net/http"
)

func NewEditUploadHandler(configService config.Service, storageService storage.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		appName, err := configService.GetAppName()
		if err != nil {
			log.Printf("Failed to get app name: %v\n", appName)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch req.Method {
		case http.MethodGet:
			q := req.URL.Query()
			i := q.Get("id")

			var fileUpload *item.FileUpload
			if i != "" {
				fileUpload, err = storageService.GetFileUpload(i)
				if err != nil {
					LogAndFail(res, err, appName)
					return
				}

				if fileUpload == nil {
					res.WriteHeader(http.StatusNotFound)
					errView, err := views.Admin(util.Html("error_404"), appName)
					if err != nil {
						return
					}

					res.Write(errView)
					return
				}
			} else {
				_, ok := interface{}(fileUpload).(item.Identifiable)
				if !ok {
					log.Println("Content type", storage.UploadsEntityName, "doesn't implement item.Identifiable")
					return
				}

				fileUpload = &item.FileUpload{}
			}

			m, err := manager.Manage(interface{}(fileUpload).(editor.Editable), storage.UploadsEntityName)
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			adminView, err := views.Admin(string(m), appName)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			res.Header().Set("Content-Type", "text/html")
			res.Write(adminView)

		case http.MethodPost:
			err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			t := req.FormValue("type")
			post, err := mappers.GetFileUploadFromFormData(req.Form)
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

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

			err = hook.BeforeSave(res, req)
			if err != nil {
				log.Println("Error running BeforeSave method in editHandler for:", t, err)
				return
			}

			// StoreFiles has the SetUpload call (which is equivalent of CreateContent in other controllers)
			files, err := mappers.GetRequestFiles(req)
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			urlPaths, err := storageService.StoreFiles(files)
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			for name, urlPath := range urlPaths {
				req.PostForm.Set(name, urlPath)
			}

			err = hook.AfterSave(res, req)
			if err != nil {
				log.Println("Error running AfterSave method in editHandler for:", t, err)
				return
			}

			scheme := req.URL.Scheme
			host := req.URL.Host
			redir := scheme + host + "/uploads"
			http.Redirect(res, req, redir, http.StatusFound)

		case http.MethodPut:
			files, err := mappers.GetRequestFiles(req)
			if err != nil {
				LogAndFail(res, err, appName)
				return
			}

			urlPaths, err := storageService.StoreFiles(files)
			if err != nil {
				log.Println("Couldn't store file uploads.", err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			res.Header().Set("Content-Type", "application/json")
			res.Write([]byte(`{"data": [{"url": "` + urlPaths["file"] + `"}]}`))
		default:
			res.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}
