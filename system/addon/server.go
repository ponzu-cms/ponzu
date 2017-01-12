package addon

import (
	"net/http"

	"github.com/ponzu-cms/ponzu/system/admin/user"
)

// Run adds Handlers to default http listener for Addon
func Run() {
	http.HandleFunc("/admin/addons", user.Auth(addonsHandler))
	http.HandleFunc("/admin/addon", user.Auth(addonHandler))
}
