package admin

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bosssauce/ponzu/system/admin/user"
)

// Run adds Handlers to default http listener for Admin
func Run() {
	http.HandleFunc("/admin", user.Auth(adminHandler))

	http.HandleFunc("/admin/init", initHandler)

	http.HandleFunc("/admin/login", loginHandler)
	http.HandleFunc("/admin/logout", logoutHandler)

	http.HandleFunc("/admin/configure", user.Auth(configHandler))
	http.HandleFunc("/admin/configure/users", user.Auth(configUsersHandler))

	http.HandleFunc("/admin/posts", user.Auth(postsHandler))
	http.HandleFunc("/admin/posts/search", user.Auth(searchHandler))

	http.HandleFunc("/admin/edit", user.Auth(editHandler))
	http.HandleFunc("/admin/edit/upload", user.Auth(editUploadHandler))

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Couldn't find current directory for file server.")
	}

	staticDir := filepath.Join(pwd, "cmd", "ponzu", "vendor", "github.com", "bosssauce", "ponzu", "system")
	http.Handle("/admin/static/", CacheControl(http.StripPrefix("/admin/", http.FileServer(http.Dir(staticDir)))))
	fmt.Println(staticDir)

	// API path needs to be registered within server package so that it is handled
	// even if the API server is not running. Otherwise, images/files uploaded
	// through the editor will not load within the admin system.
	uploadsDir := filepath.Join(pwd, "uploads")
	http.Handle("/api/uploads/", CacheControl(http.StripPrefix("/api/uploads/", http.FileServer(http.Dir(uploadsDir)))))
}
