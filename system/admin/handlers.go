package admin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bosssauce/ponzu/content"
	"github.com/bosssauce/ponzu/management/editor"
	"github.com/bosssauce/ponzu/management/manager"
	"github.com/bosssauce/ponzu/system/admin/config"
	"github.com/bosssauce/ponzu/system/admin/upload"
	"github.com/bosssauce/ponzu/system/admin/user"
	"github.com/bosssauce/ponzu/system/api"
	"github.com/bosssauce/ponzu/system/db"

	"github.com/nilslice/jwt"
	"github.com/nilslice/ponzu-test/cmd/ponzu/vendor/github.com/gorilla/schema"
)

func adminHandler(res http.ResponseWriter, req *http.Request) {
	view, err := Admin(nil)
	if err != nil {
		log.Println(err)
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
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "text/html")
		res.Write(view)

	case http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			log.Println(err)
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

		email := strings.ToLower(req.FormValue("email"))
		password := req.FormValue("password")
		usr := user.NewUser(email, password)

		_, err = db.SetUser(usr)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// set initial user email as admin_email and make config
		req.Form.Set("admin_email", email)
		err = db.SetConfig(req.Form)
		if err != nil {
			log.Println(err)
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
			Path:    "/",
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
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		c := &config.Config{}

		err = json.Unmarshal(data, c)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		cfg, err := c.MarshalEditor()
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		adminView, err := Admin(cfg)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(adminView)

	case http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = db.SetConfig(req.Form)
		if err != nil {
			log.Println(err)
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
		view, err := UsersList(req)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		res.Write(view)

	case http.MethodPost:
		// create new user
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		email := strings.ToLower(req.FormValue("email"))
		password := req.PostFormValue("password")

		if email == "" || password == "" {
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		usr := user.NewUser(email, password)

		_, err = db.SetUser(usr)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		http.Redirect(res, req, req.URL.String(), http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func configUsersEditHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB

		// check if user to be edited is current user
		j, err := db.CurrentUser(req)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		usr := &user.User{}
		err = json.Unmarshal(j, usr)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// check if password matches
		password := req.PostFormValue("password")

		if !user.IsUser(usr, password) {
			log.Println("Unexpected user/password combination for", usr.Email)
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error405()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		email := strings.ToLower(req.PostFormValue("email"))
		newPassword := req.PostFormValue("new_password")
		var updatedUser *user.User
		if newPassword != "" {
			updatedUser = user.NewUser(email, newPassword)
		} else {
			updatedUser = user.NewUser(email, password)
		}

		// set the ID to the same ID as current user
		updatedUser.ID = usr.ID

		// set user in db
		err = db.UpdateUser(usr, updatedUser)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// create new token
		week := time.Now().Add(time.Hour * 24 * 7)
		claims := map[string]interface{}{
			"exp":  week,
			"user": updatedUser.Email,
		}
		token, err := jwt.New(claims)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// add token to cookie +1 week expiration
		cookie := &http.Cookie{
			Name:    "_token",
			Value:   token,
			Expires: week,
			Path:    "/",
		}
		http.SetCookie(res, cookie)

		// add new token cookie to the request
		req.AddCookie(cookie)

		http.Redirect(res, req, strings.TrimSuffix(req.URL.String(), "/edit"), http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func configUsersDeleteHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB

		// do not allow current user to delete themselves
		j, err := db.CurrentUser(req)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		usr := &user.User{}
		err = json.Unmarshal(j, &usr)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		email := strings.ToLower(req.PostFormValue("email"))

		if usr.Email == email {
			log.Println(err)
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error405()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// delete existing user
		err = db.DeleteUser(email)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		http.Redirect(res, req, strings.TrimSuffix(req.URL.String(), "/delete"), http.StatusFound)

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
			log.Println(err)
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
			log.Println(err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		// check email & password
		j, err := db.User(strings.ToLower(req.FormValue("email")))
		if err != nil {
			log.Println(err)
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
			log.Println(err)
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
			log.Println(err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		// add it to cookie +1 week expiration
		http.SetCookie(res, &http.Cookie{
			Name:    "_token",
			Value:   token,
			Expires: week,
			Path:    "/",
		})

		http.Redirect(res, req, strings.TrimSuffix(req.URL.String(), "/login"), http.StatusFound)
	}
}

func logoutHandler(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{
		Name:    "_token",
		Expires: time.Unix(0, 0),
		Value:   "",
		Path:    "/",
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

	order := strings.ToLower(q.Get("order"))
	status := q.Get("status")

	posts := db.ContentAll(t + "_sorted")
	b := &bytes.Buffer{}

	if _, ok := content.Types[t]; !ok {
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error405()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	pt := content.Types[t]()

	p, ok := pt.(editor.Editable)
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	var hasExt bool
	_, ok = pt.(api.Externalable)
	if ok {
		hasExt = true
	}

	html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="col s8">
						<div class="row">
							<div class="card-title col s7">` + t + ` Items</div>
							<div class="col s5 input-field inline">
								<select class="browser-default __ponzu sort-order">
									<option value="DESC">New to Old</option>
									<option value="ASC">Old to New</option>
								</select>
								<label class="active">Sort:</label>
							</div>	
							<script>
								$(function() {
									var sort = $('select.__ponzu.sort-order');

									sort.on('change', function() {
										var path = window.location.pathname;
										var s = sort.val();
										var t = getParam('type');

										window.location.replace(path + '?type=' + t + '&order=' + s)
									});

									var order = getParam('order');
									if (order !== '') {
										sort.val(order);
									}
									
								});
							</script>
						</div>
					</div>
					<form class="col s4" action="/admin/posts/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all ` + t + ` fields" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
						</div>
                    </form>	
					</div>`
	if hasExt {
		if status == "" {
			q.Add("status", "public")
		}

		q.Set("status", "public")
		publicURL := req.URL.Path + "?" + q.Encode()

		q.Set("status", "pending")
		pendingURL := req.URL.Path + "?" + q.Encode()

		switch status {
		case "public", "":
			html += `<div class="row externalable">
					<span class="description">Status:</span> 
					<span class="active">Public</span>
					&nbsp;&vert;&nbsp;
					<a href="` + pendingURL + `">Pending</a>
				</div>`

		case "pending":
			// get _pending posts of type t from the db
			posts = db.ContentAll(t + "_pending")

			html += `<div class="row externalable">
					<span class="description">Status:</span> 
					<a href="` + publicURL + `">Public</a>
					&nbsp;&vert;&nbsp;
					<span class="active">Pending</span>					
				</div>`
		}

	}
	html += `<ul class="posts row">`

	switch order {
	case "desc", "":
		if hasExt {
			// reverse the order of posts slice
			for i := len(posts) - 1; i >= 0; i-- {
				err := json.Unmarshal(posts[i], &p)
				if err != nil {
					log.Println("Error unmarshal json into", t, err, posts[i])

					post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
					b.Write([]byte(post))
					continue
				}

				post := adminPostListItem(p, t, status)
				b.Write(post)
			}
		} else {
			// keep natural order of posts slice, as returned from sorted bucket
			for i := range posts {
				err := json.Unmarshal(posts[i], &p)
				if err != nil {
					log.Println("Error unmarshal json into", t, err, posts[i])

					post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
					b.Write([]byte(post))
					continue
				}

				post := adminPostListItem(p, t, status)
				b.Write(post)
			}
		}

	case "asc":
		if hasExt {
			// keep natural order of posts slice, as returned from sorted bucket
			for i := range posts {
				err := json.Unmarshal(posts[i], &p)
				if err != nil {
					log.Println("Error unmarshal json into", t, err, posts[i])

					post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
					b.Write([]byte(post))
					continue
				}

				post := adminPostListItem(p, t, status)
				b.Write(post)
			}
		} else {
			// reverse the order of posts slice
			for i := len(posts) - 1; i >= 0; i-- {
				err := json.Unmarshal(posts[i], &p)
				if err != nil {
					log.Println("Error unmarshal json into", t, err, posts[i])

					post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
					b.Write([]byte(post))
					continue
				}

				post := adminPostListItem(p, t, status)
				b.Write(post)
			}
		}
	}

	b.Write([]byte(`</ul></div></div>`))

	script := `
	<script>
		$(function() {
			var del = $('.quick-delete-post.__ponzu span');
			del.on('click', function(e) {
				if (confirm("[Ponzu] Please confirm:\n\nAre you sure you want to delete this post?\nThis cannot be undone.")) {
					$(e.target).parent().submit();
				}
			});
		});
	</script>
	`

	btn := `<div class="col s3"><a href="/admin/edit?type=` + t + `" class="btn new-post waves-effect waves-light">New ` + t + `</a></div></div>`
	html = html + b.String() + script + btn

	adminView, err := Admin([]byte(html))
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}

// adminPostListItem is a helper to create the li containing a post.
// p is the asserted post as an Editable, t is the Type of the post.
// specifier is passed to append a name to a namespace like _pending
func adminPostListItem(p editor.Editable, t, status string) []byte {
	s, ok := p.(editor.Sortable)
	if !ok {
		log.Println("Content type", t, "doesn't implement editor.Sortable")
		post := `<li class="col s12">Error retreiving data. Your data type doesn't implement necessary interfaces.</li>`
		return []byte(post)
	}

	// use sort to get other info to display in admin UI post list
	tsTime := time.Unix(int64(s.Time()/1000), 0)
	upTime := time.Unix(int64(s.Touch()/1000), 0)
	updatedTime := upTime.Format("01/02/06 03:04 PM")
	publishTime := tsTime.Format("01/02/06")

	cid := fmt.Sprintf("%d", p.ContentID())

	if status == "public" {
		status = ""
	} else {
		status = "_" + status
	}

	post := `
			<li class="col s12">
				<a href="/admin/edit?type=` + t + `&status=` + strings.TrimPrefix(status, "_") + `&id=` + cid + `">` + p.ContentName() + `</a>
				<span class="post-detail">Updated: ` + updatedTime + `</span>
				<span class="publish-date right">` + publishTime + `</span>

				<form enctype="multipart/form-data" class="quick-delete-post __ponzu right" action="/admin/edit/delete" method="post">
					<span>Delete</span>
					<input type="hidden" name="id" value="` + cid + `" />
					<input type="hidden" name="type" value="` + t + status + `" />
				</form>
			</li>`

	return []byte(post)
}

func approvePostHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		errView, err := Error405()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	t := req.FormValue("type")
	if strings.Contains(t, "_") {
		t = strings.Split(t, "_")[0]
	}

	post := content.Types[t]()

	// check if we have a Mergeable
	m, ok := post.(api.Mergeable)
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err = dec.Decode(post, req.Form)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	// call its Approve method
	err = m.Approve(req)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	// Store the content in the bucket t
	id, err = db.SetContent(t+":-1", req.Form)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	// redirect to the new approved content's editor
	redir := req.URL.Scheme + req.URL.Host + strings.TrimSuffix(req.URL.Path, "/approve")
	redir += fmt.Sprintf("?type=%s&id=%d", t, id)
	http.Redirect(res, req, http.StatusFound)
}

func editHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		q := req.URL.Query()
		i := q.Get("id")
		t := q.Get("type")
		status := q.Get("status")

		contentType, ok := content.Types[t]
		if !ok {
			fmt.Fprintf(res, content.ErrTypeNotRegistered, t)
			return
		}
		post := contentType()

		if i != "" {
			if status == "pending" {
				t = t + "_pending"
			}

			data, err := db.Content(t + ":" + i)
			if err != nil {
				log.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				errView, err := Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			if len(data) < 1 || data == nil {
				fmt.Println(string(data))
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
				log.Println(err)
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
			log.Println(err)
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
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(adminView)

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println(err)
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
		up := req.FormValue("updated")

		// create a timestamp if one was not set
		if ts == "" {
			ts := fmt.Sprintf("%d", time.Now().Unix()*1000)
			req.PostForm.Set("timestamp", ts)
		}

		if up == "" {
			req.PostForm.Set("updated", ts)
		}

		urlPaths, err := upload.StoreFiles(req)
		if err != nil {
			log.Println(err)
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
			log.Println(err)
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
		log.Println(err)
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
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// catch specifier suffix from delete form value
	if strings.Contains(t, "_") {
		spec := strings.Split(t, "_")
		t = spec[0]
	}

	redir := strings.TrimSuffix(req.URL.Scheme+req.URL.Host+req.URL.Path, "/edit/delete")
	redir = redir + "/posts?type=" + t
	http.Redirect(res, req, redir, http.StatusFound)
}

func editUploadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	urlPaths, err := upload.StoreFiles(req)
	if err != nil {
		log.Println("Couldn't store file uploads.", err)
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

		err := json.Unmarshal(posts[i], &p)
		if err != nil {
			log.Println("Error unmarshal search result json into", t, err, posts[i])

			post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
			b.Write([]byte(post))
			continue
		}

		post := adminPostListItem(p, t, "")
		b.Write([]byte(post))
	}

	b.Write([]byte(`</ul></div></div>`))

	btn := `<div class="col s3"><a href="/admin/edit?type=` + t + `" class="btn new-post waves-effect waves-light">New ` + t + `</a></div></div>`
	html = html + b.String() + btn

	adminView, err := Admin([]byte(html))
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}
