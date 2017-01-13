package api

import (
	"fmt"
	"net/http"
)

// Run adds Handlers to default http listener for API
func Run() {
	http.HandleFunc("/api/types", Record(CORS(typesHandler)))

	http.HandleFunc("/api/contents", Record(CORS(contentsHandler)))

	http.HandleFunc("/api/content", Record(CORS(contentHandler)))

	http.HandleFunc("/api/content/external", Record(CORS(externalContentHandler)))

	fmt.Println("API routes registered.")
}
