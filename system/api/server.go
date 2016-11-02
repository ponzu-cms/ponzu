package api

import (
	"net/http"
)

// Run adds Handlers to default http listener for API
func Run() {
	http.HandleFunc("/api/types", CORS(Record(typesHandler)))

	http.HandleFunc("/api/posts", CORS(Record(postsHandler)))

	http.HandleFunc("/api/post", CORS(Record(postHandler)))

	http.HandleFunc("/api/external/posts", CORS(Record(externalPostsHandler)))
}
