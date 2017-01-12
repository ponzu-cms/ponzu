package addon

import (
	"bytes"
	"log"
	"net/http"
	"strings"

	"github.com/ponzu-cms/ponzu/system/admin"
	"github.com/ponzu-cms/ponzu/system/db"
	"github.com/ponzu-cms/ponzu/system/item"

	"github.com/tidwall/gjson"
)

func addonsHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		all := db.AddonAll()
		list := &bytes.Buffer{}

		for i := range all {
			v := adminAddonListItem(all[i])
			_, err := list.Write(v)
			if err != nil {
				log.Println("Error writing bytes to addon list view:", err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := admin.Error500()
				if err != nil {
					log.Println(err)
					return
				}

				res.Write(errView)
				return
			}
		}

		html := &bytes.Buffer{}
		open := `<div class="col s9 card">		
				<div class="card-content">
				<div class="row">
				<div class="card-title col s7">Addons</div>	
				</div>
				<ul class="posts row">`

		_, err := html.WriteString(open)
		if err != nil {
			log.Println("Error writing open html to addon html view:", err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := admin.Error500()
			if err != nil {
				log.Println(err)
				return
			}

			res.Write(errView)
			return
		}

		_, err = html.Write(list.Bytes())
		if err != nil {
			log.Println("Error writing list bytes to addon html view:", err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := admin.Error500()
			if err != nil {
				log.Println(err)
				return
			}

			res.Write(errView)
			return
		}

		_, err = html.WriteString(`</ul></div></div>`)
		if err != nil {
			log.Println("Error writing close html to addon html view:", err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := admin.Error500()
			if err != nil {
				log.Println(err)
				return
			}

			res.Write(errView)
			return
		}

		if html.Len() == 0 {
			_, err := html.WriteString(`<p>No addons available.</p>`)
			if err != nil {
				log.Println("Error writing default addon html to admin view:", err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := admin.Error500()
				if err != nil {
					log.Println(err)
					return
				}

				res.Write(errView)
				return
			}
		}

		view, err := admin.Admin(html.Bytes())
		if err != nil {
			log.Println("Error writing addon html to admin view:", err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := admin.Error500()
			if err != nil {
				log.Println(err)
				return
			}

			res.Write(errView)
			return
		}

		res.Write(view)

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := admin.Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		id := req.PostFormValue("id")
		action := strings.ToLower(req.PostFormValue("action"))

		_, err = db.Addon(id)
		if err == db.ErrNoAddonExists {
			log.Println(err)
			res.WriteHeader(http.StatusNotFound)
			errView, err := admin.Error404()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := admin.Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		switch action {
		case "enable":
			err := Enable(id)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := admin.Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}
		case "disable":
			err := Disable(id)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := admin.Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}
		default:
			res.WriteHeader(http.StatusBadRequest)
			errView, err := admin.Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		http.Redirect(res, req, req.URL.String(), http.StatusFound)

	default:
		res.WriteHeader(http.StatusBadRequest)
		errView, err := admin.Error400()
		if err != nil {
			log.Println(err)
			return
		}

		res.Write(errView)
		return
	}
}

func addonHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		id := req.FormValue("id")

		data, err := db.Addon(id)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := admin.Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		_, ok := Types[id]
		if !ok {
			log.Println("Addon: ", id, "is not found in addon.Types map")
			res.WriteHeader(http.StatusNotFound)
			errView, err := admin.Error404()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		m, err := Manage(data, id)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := admin.Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		addonView, err := admin.Admin(m)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(addonView)

	case http.MethodPost:
		// save req.Form
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := admin.Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		name := req.FormValue("addon_name")
		id := req.FormValue("addon_reverse_dns")

		at, ok := Types[id]
		if !ok {
			log.Println("Error: addon", name, "has no record in addon.Types map at", id)
			res.WriteHeader(http.StatusBadRequest)
			errView, err := admin.Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// if Hookable, call BeforeSave prior to saving
		h, ok := at().(item.Hookable)
		if ok {
			err := h.BeforeSave(req)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := admin.Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}
		}

		err = db.SetAddon(req.Form, at())
		if err != nil {
			log.Println("Error saving addon:", name, err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := admin.Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		http.Redirect(res, req, "/admin/addon?id="+id, http.StatusFound)

	default:
		res.WriteHeader(http.StatusBadRequest)
		errView, err := admin.Error405()
		if err != nil {
			log.Println(err)
			return
		}

		res.Write(errView)
		return
	}
}

func adminAddonListItem(data []byte) []byte {
	id := gjson.GetBytes(data, "addon_reverse_dns").String()
	status := gjson.GetBytes(data, "addon_status").String()
	name := gjson.GetBytes(data, "addon_name").String()

	var action string
	var buttonClass string
	if status != StatusEnabled {
		action = "Enable"
		buttonClass = "green"
	} else {
		action = "Disable"
		buttonClass = "red"
	}

	a := `
			<li class="col s12">
				<a href="/admin/addon?id=` + id + `" alt="Configure '` + name + `'">` + name + `</a>

				<form enctype="multipart/form-data" class="quick-` + strings.ToLower(action) + `-addon __ponzu right" action="/admin/addons" method="post">
					<button class="btn waves-effect waves-effect-light ` + buttonClass + `">` + action + `</button>
					<input type="hidden" name="id" value="` + id + `" />
					<input type="hidden" name="action" value="` + action + `" />
				</form>
			</li>`

	return []byte(a)
}
