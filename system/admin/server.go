package admin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nilslice/cms/content"
	"github.com/nilslice/cms/management/editor"
	"github.com/nilslice/cms/management/manager"
	"github.com/nilslice/cms/system/admin/config"
	"github.com/nilslice/cms/system/admin/user"
	"github.com/nilslice/cms/system/db"
	"github.com/nilslice/jwt"
)

// Run adds Handlers to default http listener for Admin
func Run() {
	http.HandleFunc("/admin", user.Auth(adminHandler))

	http.HandleFunc("/admin/static/", func(res http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		pathParts := strings.Split(path, "/")[1:]
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Coudln't get current directory to set static asset source.")
		}

		filePathParts := make([]string, len(pathParts)+2, len(pathParts)+2)
		filePathParts = append(filePathParts, pwd)
		filePathParts = append(filePathParts, "system")
		filePathParts = append(filePathParts, pathParts...)

		http.ServeFile(res, req, filepath.Join(filePathParts...))
	})

	http.HandleFunc("/admin/init", func(res http.ResponseWriter, req *http.Request) {
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

			err = db.SetConfig(req.Form)
			if err != nil {
				fmt.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			email := req.FormValue("email")
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
	})

	http.HandleFunc("/admin/login", loginHandler)
	http.HandleFunc("/admin/logout", logoutHandler)

	http.HandleFunc("/admin/configure", func(res http.ResponseWriter, req *http.Request) {
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

	})

	http.HandleFunc("/admin/configure/users", func(res http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			// list all users and delete buttons

		case http.MethodPost:
			// create new user

		default:
			res.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/admin/posts", func(res http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()
		t := q.Get("type")
		if t == "" {
			res.WriteHeader(http.StatusBadRequest)
		}

		posts := db.ContentAll(t)
		b := &bytes.Buffer{}
		p := content.Types[t]().(editor.Editable)

		html := `<div class="col s9 card">		
					<div class="card-content row">
					<div class="card-title">` + t + ` Items</div>		
					<ul class="posts">`

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
				data, err := db.Content(t + ":" + i)
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
			if err != nil {
				fmt.Println(err)
				res.WriteHeader(http.StatusInternalServerError)
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
	})

	http.HandleFunc("/admin/edit/upload", func(res http.ResponseWriter, req *http.Request) {
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
	})

	// API path needs to be registered within server package so that it is handled
	// even if the API server is not running. Otherwise, images/files uploaded
	// through the editor will not load within the admin system.
	http.HandleFunc("/api/uploads/", func(res http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		pathParts := strings.Split(path, "/")[2:]

		pwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Coudln't get current directory to set static asset source.")
		}

		filePathParts := make([]string, len(pathParts)+1, len(pathParts)+1)
		filePathParts = append(filePathParts, pwd)
		filePathParts = append(filePathParts, pathParts...)

		http.ServeFile(res, req, filepath.Join(filePathParts...))
	})
}

func storeFileUploads(req *http.Request) (map[string]string, error) {
	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	ts := req.FormValue("timestamp")

	// To use for FormValue name:urlPath
	urlPaths := make(map[string]string)

	// get ts values individually to use as directory names when storing
	// uploaded images
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
	} else {
		tsParts := strings.Split(ts, "-")
		year, err := strconv.Atoi(tsParts[0])
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}

		month, err := strconv.Atoi(tsParts[1])
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}

		day, err := strconv.Atoi(tsParts[2])
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}

		date["year"] = year
		date["month"] = month
		date["day"] = day
	}

	// get or create upload directory to save files from request
	pwd, err := os.Getwd()
	if err != nil {
		err := fmt.Errorf("Failed to locate current directory: %s", err)
		return nil, err
	}

	tsParts := strings.Split(ts, "-")
	urlPathPrefix := "api"
	uploadDirName := "uploads"

	uploadDir := filepath.Join(pwd, uploadDirName, tsParts[0], tsParts[1])
	err = os.MkdirAll(uploadDir, os.ModeDir|os.ModePerm)

	// loop over all files and save them to disk
	for name, fds := range req.MultipartForm.File {
		filename := fds[0].Filename
		src, err := fds[0].Open()
		if err != nil {
			err := fmt.Errorf("Couldn't open uploaded file: %s", err)
			return nil, err

		}
		defer src.Close()

		// check if file at path exists, if so, add timestamp to file
		absPath := filepath.Join(uploadDir, filename)

		if _, err := os.Stat(absPath); !os.IsNotExist(err) {
			filename = fmt.Sprintf("%d-%s", time.Now().Unix(), filename)
			absPath = filepath.Join(uploadDir, filename)
		}

		// save to disk (TODO: or check if S3 credentials exist, & save to cloud)
		dst, err := os.Create(absPath)
		if err != nil {
			err := fmt.Errorf("Failed to create destination file for upload: %s", err)
			return nil, err
		}

		// copy file from src to dst on disk
		if _, err = io.Copy(dst, src); err != nil {
			err := fmt.Errorf("Failed to copy uploaded file to destination: %s", err)
			return nil, err
		}

		// add name:urlPath to req.PostForm to be inserted into db
		urlPath := fmt.Sprintf("/%s/%s/%s/%s/%s", urlPathPrefix, uploadDirName, tsParts[0], tsParts[1], filename)

		urlPaths[name] = urlPath
	}

	return urlPaths, nil
}
