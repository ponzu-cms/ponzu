package admin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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

	"github.com/gorilla/schema"
	emailer "github.com/nilslice/email"
	"github.com/nilslice/jwt"
)

func adminHandler(res http.ResponseWriter, req *http.Request) {
	view, err := Dashboard()
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

func forgotPasswordHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		view, err := ForgotPassword()
		if err != nil {
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
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		// check email for user, if no user return Error
		email := strings.ToLower(req.FormValue("email"))
		if email == "" {
			res.WriteHeader(http.StatusBadRequest)
			log.Println("Failed account recovery. No email address submitted.")
			return
		}

		_, err = db.User(email)
		if err == db.ErrNoUserExists {
			res.WriteHeader(http.StatusBadRequest)
			log.Println("No user exists.", err)
			return
		}

		if err != db.ErrNoUserExists && err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Println("Error:", err)
			return
		}

		// create temporary key to verify user
		key, err := db.SetRecoveryKey(email)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed to set account recovery key.", err)
			return
		}

		domain, err := db.Config("domain")
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed to get domain from configuration.", err)
			return
		}

		body := fmt.Sprintf(`
There has been an account recovery request made for the user with email:
%s

To recover your account, please go to http://%s/admin/recover/key and enter 
this email address along with the following secret key:

%s

If you did not make the request, ignore this message and your password 
will remain as-is.


Thank you,
Ponzu CMS at %s

`, email, domain, key, domain)

		msg := emailer.Message{
			To:      email,
			From:    fmt.Sprintf("ponzu@%s", domain),
			Subject: fmt.Sprintf("Account Recovery [%s]", domain),
			Body:    body,
		}

		go func() {
			err = msg.Send()
			if err != nil {
				log.Println("Failed to send message to:", msg.To, "about", msg.Subject, "Error:", err)
			}
		}()

		// redirect to /admin/recover/key and send email with key and URL
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/recover/key", http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		errView, err := Error405()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}
}

func recoveryKeyHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		view, err := RecoveryKey()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Write(view)

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			log.Println("Error parsing recovery key form:", err)

			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		// check for email & key match
		email := strings.ToLower(req.FormValue("email"))
		key := req.FormValue("key")

		var actual string
		if actual, err = db.RecoveryKey(email); err != nil || actual == "" {
			log.Println("Error getting recovery key from database:", err)

			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		if key != actual {
			log.Println("Bad recovery key submitted:", key)
			log.Println("Actual:", actual)

			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		// set user with new password
		password := req.FormValue("password")
		usr := &user.User{}
		u, err := db.User(email)
		if err != nil {
			log.Println("Error finding user by email:", email, err)

			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		if u == nil {
			log.Println("No user found with email:", email)

			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		err = json.Unmarshal(u, usr)
		if err != nil {
			log.Println("Error decoding user from database:", err)

			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		update := user.NewUser(email, password)
		update.ID = usr.ID

		err = db.UpdateUser(usr, update)
		if err != nil {
			log.Println("Error updating user:", err)

			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Error, please go back and try again."))
			return
		}

		// redirect to /admin/login
		redir := req.URL.Scheme + req.URL.Host + "/admin/login"
		http.Redirect(res, req, redir, http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func recoveryEditHandler(res http.ResponseWriter, req *http.Request) {

}

func contentsHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	order := strings.ToLower(q.Get("order"))
	if order != "asc" {
		order = "desc"
	}

	status := q.Get("status")

	if _, ok := content.Types[t]; !ok {
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
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

	count, err := strconv.Atoi(q.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if q.Get("count") == "" {
			count = 10
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}
	}

	offset, err := strconv.Atoi(q.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if q.Get("offset") == "" {
			offset = 0
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}
	}

	opts := db.QueryOptions{
		Count:  count,
		Offset: offset,
		Order:  order,
	}

	var specifier string
	if status == "public" {
		specifier = "__sorted"
	} else if status == "pending" {
		specifier = "__pending"
	}

	total, posts := db.Query(t+specifier, opts)
	b := &bytes.Buffer{}

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
					<form class="col s4" action="/admin/contents/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all ` + t + ` fields" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
							<input type="hidden" name="status" value="` + status + `" />
						</div>
                    </form>	
					</div>`
	if hasExt {
		if status == "" {
			q.Set("status", "public")
		}

		// always start from top of results when changing public/pending
		q.Del("count")
		q.Del("offset")

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

		case "pending":
			// get __pending posts of type t from the db
			_, posts = db.Query(t+"__pending", opts)

			html += `<div class="row externalable">
					<span class="description">Status:</span> 
					<a href="` + publicURL + `">Public</a>
					&nbsp;&vert;&nbsp;
					<span class="active">Pending</span>					
				</div>`

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

	} else {
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

	html += `<ul class="posts row">`

	b.Write([]byte(`</ul>`))

	statusDisabled := "disabled"
	prevStatus := ""
	nextStatus := ""
	// total may be less than 10 (default count), so reset count to match total
	if total < count {
		count = total
	}
	// nothing previous to current list
	if offset == 0 {
		prevStatus = statusDisabled
	}
	// nothing after current list
	if (offset+1)*count >= total {
		nextStatus = statusDisabled
	}

	// set up pagination values
	urlFmt := req.URL.Path + "?count=%d&offset=%d&&order=%s&status=%s&type=%s"
	prevURL := fmt.Sprintf(urlFmt, count, offset-1, order, status, t)
	nextURL := fmt.Sprintf(urlFmt, count, offset+1, order, status, t)
	start := 1 + count*offset
	end := start + count - 1

	if total < end {
		end = total
	}

	pagination := fmt.Sprintf(`
	<ul class="pagination row">
		<li class="col s2 waves-effect %s"><a href="%s"><i class="material-icons">chevron_left</i></a></li>
		<li class="col s8">%d to %d of %d</li>
		<li class="col s2 waves-effect %s"><a href="%s"><i class="material-icons">chevron_right</i></a></li>
	</ul>
	`, prevStatus, prevURL, start, end, total, nextStatus, nextURL)

	// show indicator that a collection of items will be listed implicitly, but
	// that none are created yet
	if total < 1 {
		pagination = `
		<ul class="pagination row">
			<li class="col s2 waves-effect disabled"><a href="#"><i class="material-icons">chevron_left</i></a></li>
			<li class="col s8">0 to 0 of 0</li>
			<li class="col s2 waves-effect disabled"><a href="#"><i class="material-icons">chevron_right</i></a></li>
		</ul>
		`
	}

	b.Write([]byte(pagination + `</div></div>`))

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

		// disable link from being clicked if parent is 'disabled'
		$(function() {
			$('ul.pagination li.disabled a').on('click', function(e) {
				e.preventDefault();
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
// specifier is passed to append a name to a namespace like __pending
func adminPostListItem(e editor.Editable, typeName, status string) []byte {
	s, ok := e.(content.Sortable)
	if !ok {
		log.Println("Content type", typeName, "doesn't implement content.Sortable")
		post := `<li class="col s12">Error retreiving data. Your data type doesn't implement necessary interfaces. (content.Sortable)</li>`
		return []byte(post)
	}

	i, ok := e.(content.Identifiable)
	if !ok {
		log.Println("Content type", typeName, "doesn't implement content.Identifiable")
		post := `<li class="col s12">Error retreiving data. Your data type doesn't implement necessary interfaces. (content.Identifiable)</li>`
		return []byte(post)
	}

	// use sort to get other info to display in admin UI post list
	tsTime := time.Unix(int64(s.Time()/1000), 0)
	upTime := time.Unix(int64(s.Touch()/1000), 0)
	updatedTime := upTime.Format("01/02/06 03:04 PM")
	publishTime := tsTime.Format("01/02/06")

	cid := fmt.Sprintf("%d", i.ItemID())

	switch status {
	case "public", "":
		status = ""
	default:
		status = "__" + status
	}

	post := `
			<li class="col s12">
				<a href="/admin/edit?type=` + typeName + `&status=` + strings.TrimPrefix(status, "__") + `&id=` + cid + `">` + i.String() + `</a>
				<span class="post-detail">Updated: ` + updatedTime + `</span>
				<span class="publish-date right">` + publishTime + `</span>

				<form enctype="multipart/form-data" class="quick-delete-post __ponzu right" action="/admin/edit/delete" method="post">
					<span>Delete</span>
					<input type="hidden" name="id" value="` + cid + `" />
					<input type="hidden" name="type" value="` + typeName + status + `" />
				</form>
			</li>`

	return []byte(post)
}

func approveContentHandler(res http.ResponseWriter, req *http.Request) {
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
	if strings.Contains(t, "__") {
		t = strings.Split(t, "__")[0]
	}

	post := content.Types[t]()

	// run hooks
	hook, ok := post.(content.Hookable)
	if !ok {
		log.Println("Type", t, "does not implement content.Hookable or embed content.Item.")
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	// check if we have a Mergeable
	m, ok := post.(editor.Mergeable)
	if !ok {
		log.Println("Content type", t, "must implement editor.Mergable before it can bee approved.")
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
		log.Println("Error decoding post form for content approval:", t, err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	err = hook.BeforeApprove(req)
	if err != nil {
		log.Println("Error running BeforeApprove hook in approveContentHandler for:", t, err)
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
		log.Println("Error running Approve method in approveContentHandler for:", t, err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	err = hook.AfterApprove(req)
	if err != nil {
		log.Println("Error running AfterApprove hook in approveContentHandler for:", t, err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	err = hook.BeforeSave(req)
	if err != nil {
		log.Println("Error running BeforeSave hook in approveContentHandler for:", t, err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	// Store the content in the bucket t
	id, err := db.SetContent(t+":-1", req.Form)
	if err != nil {
		log.Println("Error storing content in approveContentHandler for:", t, err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	err = hook.AfterSave(req)
	if err != nil {
		log.Println("Error running AfterSave hook in approveContentHandler for:", t, err)
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
	http.Redirect(res, req, redir, http.StatusFound)
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
				t = t + "__pending"
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
			item, ok := post.(content.Identifiable)
			if !ok {
				log.Println("Content type", t, "doesn't implement content.Identifiable")
				return
			}
			item.SetItemID(-1)
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
			req.PostForm.Set(name, urlPath)
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

		if strings.Contains(t, "__") {
			t = strings.Split(t, "__")[0]
		}

		p, ok := content.Types[t]
		if !ok {
			log.Println("Type", t, "is not a content type. Cannot edit or save.")
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		post := p()
		hook, ok := post.(content.Hookable)
		if !ok {
			log.Println("Type", t, "does not implement content.Hookable or embed content.Item.")
			res.WriteHeader(http.StatusBadRequest)
			errView, err := Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		err = hook.BeforeSave(req)
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

		err = hook.AfterSave(req)
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
		redir := scheme + host + path + "?type=" + t + "&id=" + sid

		if req.URL.Query().Get("status") == "pending" {
			redir += "&status=pending"
		}

		http.Redirect(res, req, redir, http.StatusFound)

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
	ct := t

	if id == "" || t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// catch specifier suffix from delete form value
	if strings.Contains(t, "__") {
		spec := strings.Split(t, "__")
		ct = spec[0]
	}

	p, ok := content.Types[ct]
	if !ok {
		log.Println("Type", t, "does not implement content.Hookable or embed content.Item.")
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	post := p()
	hook, ok := post.(content.Hookable)
	if !ok {
		log.Println("Type", t, "does not implement content.Hookable or embed content.Item.")
		res.WriteHeader(http.StatusBadRequest)
		errView, err := Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	reject := req.URL.Query().Get("reject")
	if reject == "true" {
		err = hook.BeforeReject(req)
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
	}

	err = hook.BeforeDelete(req)
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

	err = db.DeleteContent(t+":"+id, req.Form)
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = hook.AfterDelete(req)
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

	if reject == "true" {
		err = hook.AfterReject(req)
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
	}

	redir := strings.TrimSuffix(req.URL.Scheme+req.URL.Host+req.URL.Path, "/edit/delete")
	redir = redir + "/contents?type=" + ct
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
	status := q.Get("status")
	var specifier string

	if t == "" || search == "" {
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		return
	}

	if status == "pending" {
		specifier = "__" + status
	}

	posts := db.ContentAll(t + specifier)
	b := &bytes.Buffer{}
	p := content.Types[t]().(editor.Editable)

	html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="card-title col s7">` + t + ` Results</div>	
					<form class="col s4" action="/admin/contents/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all ` + t + ` fields" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
							<input type="hidden" name="status" value="` + status + `" />
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

		post := adminPostListItem(p, t, status)
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
