package api

import (
	"log"
	"net/http"

	"github.com/bosssauce/ponzu/content"
	"github.com/bosssauce/ponzu/system/db"
)

// Externalable accepts or rejects external POST requests to /external/posts?type=Review
type Externalable interface {
	Accept() bool
}

func externalPostsHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	t := req.URL.Query().Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	p, found := content.Types[t]
	if !found {
		log.Println("Attempt to submit content", t, "by", req.RemoteAddr)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	post := p()

	ext, ok := post.(Externalable)
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if ext.Accept() {
		db.SetContent(t+"_external"+":-1", req.Form)
	}
}
