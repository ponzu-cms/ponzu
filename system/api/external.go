package api

import (
	"log"
	"net/http"

	"github.com/bosssauce/ponzu/content"
	"github.com/bosssauce/ponzu/system/db"
)

// Externalable accepts or rejects external POST requests to /external/posts?type=Review
type Externalable interface {
	Accepts() bool
}

func externalPostsHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		log.Println("[External]", err)
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
		log.Println("[External] Attempt to submit content", t, "by", req.RemoteAddr)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	post := p()

	ext, ok := post.(Externalable)
	if !ok {
		log.Println("[External]", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if ext.Accepts() {
		_, err := db.SetContent(t+"_external"+":-1", req.Form)
		if err != nil {
			log.Println("[External]", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
