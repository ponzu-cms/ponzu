package external

import (
	"fmt"
	"net/http"

	"github.com/bosssauce/ponzu/content"
	"github.com/bosssauce/ponzu/system/api"
	"github.com/bosssauce/ponzu/system/db"
)

// Externalable accepts or rejects external POST requests to /external/posts?type=Review
type Externalable interface {
	Accept() bool
}

func init() {
	http.HandleFunc("/api/external/posts", api.CORS(externalPostsHandler))
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
		fmt.Println(t, content.Types, p)
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
