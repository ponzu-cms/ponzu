package content

import (
	"fmt"
	"log"

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

// BeforeAPIDelete is only called if the Song type implements api.Deleteable
// It is called before Delete, and returning an error will cancel the request
// causing the system to reject the data sent in the POST
func (s *Song) BeforeAPIDelete(res http.ResponseWriter, req *http.Request) error {
	// do initial user authentication here on the request, checking for a
	// token or cookie, or that certain form fields are set and valid

	// for example, this will check if the request was made by a CMS admin user:
	if !user.IsValid(req) {
		return api.ErrNoAuth
	}

	// you could then to data validation on the request post form, or do it in
	// the Delete method, which is called after BeforeAPIDelete

	return nil
}

// Delete is called after BeforeAPIDelete and implements api.Deleteable. All
// other delete-based hooks are only called if this is implemented.
func (s *Song) Delete(res http.ResponseWriter, req *http.Request) error {
	// See BeforeAPIDelete above, how we have checked the request for some
	// form of auth. This could be done here instead, but if it is done once
	// above, it means the request is valid here too.
	return nil
}

// AfterAPIDelete is called after Delete, and is useful for logging or triggering
// notifications, etc. after the data is deleted frm the database, etc.
func (s *Song) AfterAPIDelete(res http.ResponseWriter, req *http.Request) error {
	addr := req.RemoteAddr
	log.Println("Song deleted by:", addr, "id:", req.URL.Query().Get("id"))

	return nil
}
