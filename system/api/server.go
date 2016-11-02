package api

import (
	"net/http"

	a "github.com/bosssauce/ponzu/system/api/analytics"
)

// Run adds Handlers to default http listener for API
func Run() {
	http.HandleFunc("/api/types", CORS(a.Record(typesHandler)))

	http.HandleFunc("/api/posts", CORS(a.Record(postsHandler)))

	http.HandleFunc("/api/post", CORS(a.Record(postHandler)))

	http.HandleFunc("/api/external/posts", CORS(a.Record(externalPostsHandler)))
}
