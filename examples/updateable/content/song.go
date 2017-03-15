package content

import (
	"fmt"
	"log"
	"strings"

	"net/http"

	"github.com/ponzu-cms/ponzu/management/editor"
	"github.com/ponzu-cms/ponzu/system/admin/user"
	"github.com/ponzu-cms/ponzu/system/api"
	"github.com/ponzu-cms/ponzu/system/item"
)

type Song struct {
	item.Item

	Title      string `json:"title"`
	Artist     string `json:"artist"`
	Rating     int    `json:"rating"`
	Opinion    string `json:"opinion"`
	SpotifyURL string `json:"spotify_url"`
}

// MarshalEditor writes a buffer of html to edit a Song within the CMS
// and implements editor.Editable
func (s *Song) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(s,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Song field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Title", s, map[string]string{
				"label":       "Title",
				"type":        "text",
				"placeholder": "Enter the Title here",
			}),
		},
		editor.Field{
			View: editor.Input("Artist", s, map[string]string{
				"label":       "Artist",
				"type":        "text",
				"placeholder": "Enter the Artist here",
			}),
		},
		editor.Field{
			View: editor.Input("Rating", s, map[string]string{
				"label":       "Rating",
				"type":        "text",
				"placeholder": "Enter the Rating here",
			}),
		},
		editor.Field{
			View: editor.Richtext("Opinion", s, map[string]string{
				"label":       "Opinion",
				"placeholder": "Enter the Opinion here",
			}),
		},
		editor.Field{
			View: editor.Input("SpotifyURL", s, map[string]string{
				"label":       "SpotifyURL",
				"type":        "text",
				"placeholder": "Enter the SpotifyURL here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Song editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Song"] = func() interface{} { return new(Song) }
}

// String defines the display name of a Song in the CMS list-view
func (s *Song) String() string { return s.Title }

// BeforeAPIUpdate is only called if the Song type implements api.Updateable
// It is called before Update, and returning an error will cancel the request
// causing the system to reject the data sent in the POST
func (s *Song) BeforeAPIUpdate(res http.ResponseWriter, req *http.Request) error {
	// do initial user authentication here on the request, checking for a
	// token or cookie, or that certain form fields are set and valid

	// for example, this will check if the request was made by a CMS admin user:
	if !user.IsValid(req) {
		return api.ErrNoAuth
	}

	// you could then to data validation on the request post form, or do it in
	// the Update method, which is called after BeforeAPIUpdate

	return nil
}

// Update is called after BeforeAPIUpdate and is where you may influence the
// merge process.  For example, maybe you don't want an empty string for the Title
// or Artist field to be accepted by the update request.  Updates will always merge
// with existing values, but by default will accept zero value as an update if sent.
func (s *Song) Update(res http.ResponseWriter, req *http.Request) error {
	addr := req.RemoteAddr
	log.Println("Song update sent by:", addr, "id:", req.URL.Query().Get("id"))

	// On update its fine if fields are missing, but we don't want
	// title overwritten by a blank or empty string since that would
	// break the display name.  Artist is also required to be non-blank.
	var required = map[string]interface{}{
		"title":  nil,
		"artist": nil,
	}

	for k, _ := range req.PostForm {
		blank := (strings.TrimSpace(req.PostFormValue(k)) == "")
		if _, ok := required[k]; ok && blank {
			log.Println("Removing blank value for:", k)
			// We'll just remove the blank values.
			// Alternately we could return an error to
			// reject the post.
			req.PostForm.Del(k)
		}
	}

	return nil
}

// AfterAPIUpdate is called after Update, and is useful for logging or triggering
// notifications, etc. after the data is saved to the database, etc.
// The request has a context containing the databse 'target' affected by the
// request.
func (s *Song) AfterAPIUpdate(res http.ResponseWriter, req *http.Request) error {
	addr := req.RemoteAddr
	log.Println("Song updated by:", addr, "id:", req.URL.Query().Get("id"))

	return nil
}
