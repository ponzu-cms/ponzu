package api

import (
	"net/http"

	"github.com/bosssauce/ponzu/system/db"
)

// Run adds Handlers to default http listener for API
func Run() {
	http.HandleFunc("/api/types", db.CacheControl(CORS(Record(typesHandler))))

	http.HandleFunc("/api/contents", db.CacheControl(CORS(Record(contentsHandler))))

	http.HandleFunc("/api/content", db.CacheControl(CORS(Record(contentHandler))))

	http.HandleFunc("/api/content/external", db.CacheControl(CORS(Record(externalContentHandler))))
}
