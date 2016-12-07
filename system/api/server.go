package api

import (
	"net/http"

	"github.com/bosssauce/ponzu/system/db"
)

// Run adds Handlers to default http listener for API
func Run() {
	http.HandleFunc("/api/types", Record(db.CacheControl(CORS(typesHandler))))

	http.HandleFunc("/api/contents", Record(db.CacheControl(CORS(contentsHandler))))

	http.HandleFunc("/api/content", Record(db.CacheControl(CORS(contentHandler))))

	http.HandleFunc("/api/content/external", Record(db.CacheControl(CORS(externalContentHandler))))
}
