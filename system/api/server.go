package api

import "net/http"

// Run adds Handlers to default http listener for API
func Run() {
	http.HandleFunc("/api/types", CORS(typesHandler))

	http.HandleFunc("/api/posts", CORS(postsHandler))

	http.HandleFunc("/api/post", CORS(postHandler))
}
