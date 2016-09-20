package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nilslice/cms/content"
	"github.com/nilslice/cms/management/editor"
	"github.com/nilslice/cms/management/manager"
	"github.com/nilslice/cms/system/db"
)

func main() {
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
				fmt.Println("Need to show post id:", i, "(", t, ")")

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
			}

			view, err := manager.Manage(post.(editor.Editable), t)
			if err != nil {
				fmt.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "text/html")
			res.Write(view)

		case http.MethodPost:
			err := req.ParseForm()
			if err != nil {
				fmt.Println(err)
				res.WriteHeader(http.StatusBadRequest)
				return
			}

			cid := req.FormValue("id")
			t := req.FormValue("type")
			fmt.Println("query data: t=", t, "id=", cid)

			id, err := db.Set(t+":"+cid, req.PostForm)
			if err != nil {
				fmt.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			fmt.Println(t, "post created:", id)
			scheme := req.URL.Scheme
			host := req.URL.Host
			path := req.URL.Path
			desURL := scheme + host + path + "?type=" + t + "&id=" + fmt.Sprintf("%d", id)
			http.Redirect(res, req, desURL, http.StatusFound)
		}
	})

	http.ListenAndServe(":8080", nil)

}
