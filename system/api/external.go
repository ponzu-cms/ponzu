package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bosssauce/ponzu/content"
	"github.com/bosssauce/ponzu/system/admin/upload"
	"github.com/bosssauce/ponzu/system/db"
)

// Externalable accepts or rejects external POST requests to endpoints such as:
// /external/content?type=Review
type Externalable interface {
	// Accept allows external content submissions of a specific type
	Accept(req *http.Request) error
}

// Trustable allows external content to be auto-approved, meaning content sent
// as an Externalable will be stored in the public content bucket
type Trustable interface {
	AutoApprove(req *http.Request) error
}

func externalPostHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		log.Println("[External] error:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	t := req.URL.Query().Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	p, found := content.Types[t]
	if !found {
		log.Println("[External] attempt to submit unknown type:", t, "from:", req.RemoteAddr)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	post := p()

	ext, ok := post.(Externalable)
	if !ok {
		log.Println("[External] rejected non-externalable type:", t, "from:", req.RemoteAddr)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	ts := fmt.Sprintf("%d", time.Now().Unix()*1000)
	req.PostForm.Set("timestamp", ts)
	req.PostForm.Set("updated", ts)

	urlPaths, err := upload.StoreFiles(req)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	for name, urlPath := range urlPaths {
		req.PostForm.Add(name, urlPath)
	}

	// check for any multi-value fields (ex. checkbox fields)
	// and correctly format for db storage. Essentially, we need
	// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
	var discardKeys []string
	for k, v := range req.PostForm {
		if strings.Contains(k, ".") {
			key := strings.Split(k, ".")[0]

			if req.PostForm.Get(key) == "" {
				req.PostForm.Set(key, v[0])
				discardKeys = append(discardKeys, k)
			} else {
				req.PostForm.Add(key, v[0])
			}
		}
	}

	for _, discardKey := range discardKeys {
		req.PostForm.Del(discardKey)
	}

	// call Accept with the request, enabling developer to add or chack data
	// before saving to DB
	err = ext.Accept(req)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	hook, ok := post.(content.Hookable)
	if !ok {
		log.Println("[External] error: Type", t, "does not implement content.Hookable or embed content.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	err = hook.BeforeSave(req)
	if err != nil {
		log.Println("[External] error:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// set specifier for db bucket in case content is/isn't Trustable
	var spec string

	// check if the content is Trustable should be auto-approved
	trusted, ok := post.(Trustable)
	if ok {
		err := trusted.AutoApprove(req)
		if err != nil {
			log.Println("[External] error:", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		spec = "_pending"
	}

	_, err = db.SetContent(t+spec+":-1", req.PostForm)
	if err != nil {
		log.Println("[External] error:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = hook.AfterSave(req)
	if err != nil {
		log.Println("[External] error:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

}
