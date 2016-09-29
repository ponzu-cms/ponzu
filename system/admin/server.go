package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nilslice/cms/content"
	"github.com/nilslice/cms/management/editor"
	"github.com/nilslice/cms/management/manager"
	"github.com/nilslice/cms/system/db"
)

func init() {
	http.HandleFunc("/admin", func(res http.ResponseWriter, req *http.Request) {
		adminView := Admin(nil)

		res.Header().Set("Content-Type", "text/html")
		res.Write(adminView)
	})

	http.HandleFunc("/admin/posts", func(res http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()
		t := q.Get("type")
		if t == "" {
			res.WriteHeader(http.StatusBadRequest)
		}

		posts := db.GetAll(t)
		b := &bytes.Buffer{}
		p := content.Types[t]().(editor.Editable)

		html := `<a href="/admin/edit?type=` + t + `" class="button">New ` + t + `</a>
			<ul class="posts">`
		for i := range posts {
			json.Unmarshal(posts[i], &p)
			post := `<li><a href="/admin/edit?type=` +
				t + `&id=` + fmt.Sprintf("%d", p.ContentID()) +
				`">` + p.ContentName() + `</a></li>`
			b.Write([]byte(post))
		}
		html = html + b.String()

		adminView := Admin([]byte(html))

		res.Header().Set("Content-Type", "text/html")
		res.Write(adminView)
	})

	http.HandleFunc("/admin/edit", func(res http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			q := req.URL.Query()
			i := q.Get("id")
			t := q.Get("type")
			contentType, ok := content.Types[t]
			if !ok {
				fmt.Fprintf(res, content.ErrTypeNotRegistered, t)
				return
			}
			post := contentType()

			if i != "" {
				data, err := db.Get(t + ":" + i)
				if err != nil {
					fmt.Println(err)
					res.WriteHeader(http.StatusInternalServerError)
					return
				}

				err = json.Unmarshal(data, post)
				if err != nil {
					fmt.Println(err)
					res.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				post.(editor.Editable).SetContentID(-1)
			}

			m, err := manager.Manage(post.(editor.Editable), t)
			adminView := Admin(m)
			if err != nil {
				fmt.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "text/html")
			res.Write(adminView)

		case http.MethodPost:
			err := req.ParseForm()
			if err != nil {
				fmt.Println(err)
				res.WriteHeader(http.StatusBadRequest)
				return
			}

			cid := req.FormValue("id")
			t := req.FormValue("type")

			// check for any multi-value fields (ex. checkbox fields)
			// and correctly format for db storage. Essentially, we need
			// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
			var discardKeys []string
			for k, v := range req.PostForm {
				if strings.Contains(k, ".") {
					key := strings.Split(k, ".")[0]

					if req.PostForm.Get(key) == "" {
						req.PostForm.Set(key, v[0])
						discardKeys = append(discardKeys, k)
					} else {
						req.PostForm.Add(key, v[0])
					}
				}
			}

			for _, discardKey := range discardKeys {
				req.PostForm.Del(discardKey)
			}

			id, err := db.Set(t+":"+cid, req.PostForm)
			if err != nil {
				fmt.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			scheme := req.URL.Scheme
			host := req.URL.Host
			path := req.URL.Path
			sid := fmt.Sprintf("%d", id)
			desURL := scheme + host + path + "?type=" + t + "&id=" + sid
			http.Redirect(res, req, desURL, http.StatusFound)
		}
	})
}

// Run starts the Admin system on the port provided
func Run(port string) {
	http.ListenAndServe(":"+port, nil)
}
