package admin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bosssauce/ponzu/content"
	"github.com/bosssauce/ponzu/management/editor"
	"github.com/bosssauce/ponzu/management/manager"
	"github.com/bosssauce/ponzu/system/admin/config"
	"github.com/bosssauce/ponzu/system/admin/user"
	"github.com/bosssauce/ponzu/system/db"

	"github.com/nilslice/jwt"
)

func adminHandler(res http.ResponseWriter, req *http.Request) {
	view, err := Admin(nil)
	if err != nil {
		fmt.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(view)
}

func initHandler(res http.ResponseWriter, req *http.Request) {
	if db.SystemInitComplete() {
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		return
	}

	switch req.Method {
	case http.MethodGet:
		view, err := Init()
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
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// get the site name from post to encode and use as secret
		name := []byte(req.FormValue("name"))
		secret := base64.StdEncoding.EncodeToString(name)
		req.Form.Set("client_secret", secret)

		// generate an Etag to use for response caching
		etag := db.NewEtag()
		req.Form.Set("etag", etag)

		err = db.SetConfig(req.Form)
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		email := strings.ToLower(req.FormValue("email"))
		password := req.FormValue("password")
		usr := user.NewUser(email, password)

		_, err = db.SetUser(usr)
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// add _token cookie for login persistence
		week := time.Now().Add(time.Hour * 24 * 7)
		claims := map[string]interface{}{
			"exp":  week.Unix(),
			"user": usr.Email,
		}

		jwt.Secret([]byte(secret))
		token, err := jwt.New(claims)

		http.SetCookie(res, &http.Cookie{
			Name:    "_token",
			Value:   token,
			Expires: week,
		})

		redir := strings.TrimSuffix(req.URL.String(), "/init")
		http.Redirect(res, req, redir, http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func configHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		data, err := db.ConfigAll()
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		c := &config.Config{}

		err = json.Unmarshal(data, c)
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		cfg, err := c.MarshalEditor()
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		adminView, err := Admin(cfg)
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
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = db.SetConfig(req.Form)
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(res, req, req.URL.String(), http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func configUsersHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		// list all users and delete buttons

	case http.MethodPost:
		// create new user

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func loginHandler(res http.ResponseWriter, req *http.Request) {
	if !db.SystemInitComplete() {
		redir := req.URL.Scheme + req.URL.Host + "/admin/init"
		http.Redirect(res, req, redir, http.StatusFound)
		return
	}

	switch req.Method {
	case http.MethodGet:
		if user.IsValid(req) {
			http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
			return
		}

		view, err := Login()
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(view)

	case http.MethodPost:
		if user.IsValid(req) {
			http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
			return
		}

		err := req.ParseForm()
		if err != nil {
			fmt.Println(err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		// check email & password
		j, err := db.User(strings.ToLower(req.FormValue("email")))
		if err != nil {
			fmt.Println(err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		if j == nil {
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		usr := &user.User{}
		err = json.Unmarshal(j, usr)
		if err != nil {
			fmt.Println(err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		if !user.IsUser(usr, req.FormValue("password")) {
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}
		// create new token
		week := time.Now().Add(time.Hour * 24 * 7)
		claims := map[string]interface{}{
			"exp":  week,
			"user": usr.Email,
		}
		token, err := jwt.New(claims)
		if err != nil {
			fmt.Println(err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		// add it to cookie +1 week expiration
		http.SetCookie(res, &http.Cookie{
			Name:    "_token",
			Value:   token,
			Expires: week,
		})

		http.Redirect(res, req, strings.TrimSuffix(req.URL.String(), "/login"), http.StatusFound)
	}
}

func logoutHandler(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{
		Name:    "_token",
		Expires: time.Unix(0, 0),
		Value:   "",
	})

	http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/login", http.StatusFound)
}

func postsHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error405()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	posts := db.ContentAll(t)
	b := &bytes.Buffer{}
	p, ok := content.Types[t]().(editor.Editable)
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="card-title col s7">` + t + ` Items</div>	
					<form class="col s5" action="/admin/posts/search" method="get">
						<div class="input-field post-search inline">
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Search for ` + t + ` content" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
						</div>
                    </form>	
					</div>
					<ul class="posts row">`

	for i := range posts {
		json.Unmarshal(posts[i], &p)
		post := `<li class="col s12"><a href="/admin/edit?type=` +
			t + `&id=` + fmt.Sprintf("%d", p.ContentID()) +
			`">` + p.ContentName() + `</a></li>`
		b.Write([]byte(post))
	}

	b.Write([]byte(`</ul></div></div>`))

	btn := `<div class="col s3"><a href="/admin/edit?type=` + t + `" class="btn new-post waves-effect waves-light">New ` + t + `</a></div></div>`
	html = html + b.String() + btn

	adminView, err := Admin([]byte(html))
	if err != nil {
		fmt.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}

func editHandler(res http.ResponseWriter, req *http.Request) {
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
			data, err := db.Content(t + ":" + i)
			if err != nil {
				fmt.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			if len(data) < 1 || data == nil {
				res.WriteHeader(http.StatusNotFound)
				errView, err := Error404()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			err = json.Unmarshal(data, post)
			if err != nil {
				fmt.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}
		} else {
			post.(editor.Editable).SetContentID(-1)
		}

		m, err := manager.Manage(post.(editor.Editable), t)
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		adminView, err := Admin(m)
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(adminView)

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error405()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		cid := req.FormValue("id")
		t := req.FormValue("type")
		ts := req.FormValue("timestamp")

		// create a timestamp if one was not set
		date := make(map[string]int)
		if ts == "" {
			now := time.Now()
			date["year"] = now.Year()
			date["month"] = int(now.Month())
			date["day"] = now.Day()

			// create timestamp format 'yyyy-mm-dd' and set in PostForm for
			// db insertion
			ts = fmt.Sprintf("%d-%02d-%02d", date["year"], date["month"], date["day"])
			req.PostForm.Set("timestamp", ts)
		}

		urlPaths, err := storeFileUploads(req)
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		for name, urlPath := range urlPaths {
			req.PostForm.Add(name, urlPath)
		}

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

		id, err := db.SetContent(t+":"+cid, req.PostForm)
		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		scheme := req.URL.Scheme
		host := req.URL.Host
		path := req.URL.Path
		sid := fmt.Sprintf("%d", id)
		desURL := scheme + host + path + "?type=" + t + "&id=" + sid
		http.Redirect(res, req, desURL, http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func deleteHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	id := req.FormValue("id")
	t := req.FormValue("type")

	if id == "" || t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	err = db.DeleteContent(t + ":" + id)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	redir := strings.TrimSuffix(req.URL.Scheme+req.URL.Host+req.URL.Path, "/edit/delete")
	redir = redir + "posts?type=" + t
	http.Redirect(res, req, redir, http.StatusFound)
}

func editUploadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	urlPaths, err := storeFileUploads(req)
	if err != nil {
		fmt.Println("Couldn't store file uploads.", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write([]byte(`{"data": [{"url": "` + urlPaths["file"] + `"}]}`))
}

func searchHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	search := q.Get("q")

	if t == "" || search == "" {
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		return
	}

	posts := db.ContentAll(t)
	b := &bytes.Buffer{}
	p := content.Types[t]().(editor.Editable)

	html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="card-title col s7">` + t + ` Results</div>	
					<form class="col s5" action="/admin/posts/search" method="get">
						<div class="input-field post-search inline">
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Search for ` + t + ` content" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
						</div>
                    </form>	
					</div>
					<ul class="posts row">`

	for i := range posts {
		// skip posts that don't have any matching search criteria
		match := strings.ToLower(search)
		all := strings.ToLower(string(posts[i]))
		if !strings.Contains(all, match) {
			continue
		}

		json.Unmarshal(posts[i], &p)
		post := `<li class="col s12"><a href="/admin/edit?type=` +
			t + `&id=` + fmt.Sprintf("%d", p.ContentID()) +
			`">` + p.ContentName() + `</a></li>`
		b.Write([]byte(post))
	}

	b.Write([]byte(`</ul></div></div>`))

	btn := `<div class="col s3"><a href="/admin/edit?type=` + t + `" class="btn new-post waves-effect waves-light">New ` + t + `</a></div></div>`
	html = html + b.String() + btn

	adminView, err := Admin([]byte(html))
	if err != nil {
		fmt.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}
