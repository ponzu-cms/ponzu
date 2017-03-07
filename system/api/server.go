package api

import "net/http"

// Run adds Handlers to default http listener for API
func Run() {
	http.HandleFunc("/api/contents", Record(CORS(Gzip(contentsHandler))))

	http.HandleFunc("/api/content", Record(CORS(Gzip(contentHandler))))

	http.HandleFunc("/api/content/external", Record(CORS(externalContentHandler)))

	http.HandleFunc("/api/content/update", Record(CORS(updateContentHandler)))
}
