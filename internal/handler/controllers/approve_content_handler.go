package controllers

import (
	"context"
	"fmt"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/content"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"github.com/fanky5g/ponzu/internal/domain/services/management/editor"
	"github.com/fanky5g/ponzu/internal/handler/controllers/views"
	"github.com/fanky5g/ponzu/internal/util"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"strings"
)

func NewApproveContentHandler(configService config.Service, contentService content.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		appName, err := configService.GetAppName()
		if err != nil {
			log.Printf("Failed to get app name: %v\n", appName)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusMethodNotAllowed)
			errView, err := views.Admin(util.Html("error_405"), appName)
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		err = req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			LogAndFail(res, err, appName)
			return
		}

		pendingID := req.FormValue("id")

		t := req.FormValue("type")

		if strings.Contains(t, "__") {
			t = strings.Split(t, "__")[0]
		}

		if strings.Contains(t, "__") {
			t = strings.Split(t, "__")[0]
		}

		post := item.Types[t]()

		// run hooks
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

		// check if we have a Mergeable
		m, ok := post.(editor.Mergeable)
		if !ok {
			log.Println("Content type", t, "must implement editor.Mergeable before it can be approved.")
			res.WriteHeader(http.StatusBadRequest)
			errView, err := views.Admin(util.Html("error_400"), appName)
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		dec.SetAliasTag("json")
		err = dec.Decode(post, req.Form)
		if err != nil {
			log.Println("Error decoding post form for content approval:", t, err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := views.Admin(util.Html("error_400"), appName)
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		err = hook.BeforeApprove(res, req)
		if err != nil {
			log.Println("Error running BeforeApprove hook in approveContentHandler for:", t, err)
			return
		}

		// call its Approval method
		err = m.Approve(res, req)
		if err != nil {
			log.Println("Error running Approve method in approveContentHandler for:", t, err)
			return
		}

		err = hook.AfterApprove(res, req)
		if err != nil {
			log.Println("Error running AfterApprove hook in approveContentHandler for:", t, err)
			return
		}

		err = hook.BeforeSave(res, req)
		if err != nil {
			log.Println("Error running BeforeSave hook in approveContentHandler for:", t, err)
			return
		}

		// Store the content in the bucket t
		id, err := contentService.CreateContent(t+":-1", req.Form)
		if err != nil {
			log.Println("Error storing content in approveContentHandler for:", t, err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := views.Admin(util.Html("error_500"), appName)
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// set the target in the context so user can get saved value from db in hook
		ctx := context.WithValue(req.Context(), "target", fmt.Sprintf("%s:%s", t, id))
		req = req.WithContext(ctx)

		err = hook.AfterSave(res, req)
		if err != nil {
			log.Println("Error running AfterSave hook in approveContentHandler for:", t, err)
			return
		}

		if pendingID != "" {
			err = contentService.DeleteContent(req.FormValue("type"), pendingID)
			if err != nil {
				log.Println("Failed to remove content after approval:", err)
			}
		}

		// redirect to the new approved content's editor
		redir := req.URL.Scheme + req.URL.Host + strings.TrimSuffix(req.URL.Path, "/approve")
		redir += fmt.Sprintf("?type=%s&id=%s", t, id)
		http.Redirect(res, req, redir, http.StatusFound)
	}
}
