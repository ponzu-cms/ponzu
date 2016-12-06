package api

import (
	"net/http"
)

// Run adds Handlers to default http listener for API
func Run() {
	http.HandleFunc("/api/types", CORS(Record(typesHandler)))

	http.HandleFunc("/api/contents", CORS(Record(contentsHandler)))

	http.HandleFunc("/api/content", CORS(Record(contentHandler)))

	http.HandleFunc("/api/content/external", CORS(Record(externalContentHandler)))
}
