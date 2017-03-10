package content

import (
	"fmt"
	"log"
	"strings"

	"net/http"

	"github.com/ponzu-cms/ponzu/management/editor"
	"github.com/ponzu-cms/ponzu/system/admin/user"
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

// Accept implements api.Externalable, and allows external POST requests from clients
// to add content as long as the request contains the json tag names of the Song
// struct fields, and is multipart encoded
func (s *Song) Accept(res http.ResponseWriter, req *http.Request) error {
	// do form data validation for required fields
	required := []string{
		"title",
		"artist",
		"rating",
		"opinion",
		"spotify_url",
	}

	for _, r := range required {
		if req.PostFormValue(r) == "" {
			err := fmt.Errorf("request missing required field: %s", r)
			return err
		}
	}

	return nil
}

// BeforeAccept is only called if the Song type implements api.Externalable
// It is called before Accept, and returning an error will cancel the request
// causing the system to reject the data sent in the POST
func (s *Song) BeforeAccept(res http.ResponseWriter, req *http.Request) error {
	// do initial user authentication here on the request, checking for a
	// token or cookie, or that certain form fields are set and valid

	// for example, this will check if the request was made by a CMS admin user:
	if !user.IsValid(req) {
		addr := req.RemoteAddr
		err := fmt.Errorf("request rejected, invalid user. IP: %s", addr)
		return err
	}

	// you could then to data validation on the request post form, or do it in
	// the Accept method, which is called after BeforeAccept

	return nil
}

// BeforeAcceptUpdate is only called if the Song type implements api.Updateable
// It is called before AcceptUpdate, and returning an error will cancel the request
// causing the system to reject the data sent in the POST
func (s *Song) BeforeAcceptUpdate(res http.ResponseWriter, req *http.Request) error {
	// do initial user authentication here on the request, checking for a
	// token or cookie, or that certain form fields are set and valid

	// for example, this will check if the request was made by a CMS admin user:
	if !user.IsValid(req) {
		addr := req.RemoteAddr
		err := fmt.Errorf("request rejected, invalid user. IP: %s", addr)
		return err
	}

	// you could then to data validation on the request post form, or do it in
	// the Accept method, which is called after BeforeAccept

	return nil
}

// AcceptUpdate is called after BeforeAccept and is where you may influence the
// merge process.  For example, maybe you don't want an empty string for the Title
// or Artist field to be accepted by the update request.  Updates will always merge
// with existing values, but by default will accept zero value as an update if sent.
func (s *Song) AcceptUpdate(res http.ResponseWriter, req *http.Request) error {
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

// AfterAccept is called after Accept, and is useful for logging or triggering
// notifications, etc. after the data is saved to the database, etc.
// The request has a context containing the databse 'target' affected by the
// request. Ex. Song__pending:3 or Song:8 depending if Song implements api.Trustable
func (s *Song) AfterAccept(res http.ResponseWriter, req *http.Request) error {
	addr := req.RemoteAddr
	log.Println("Song sent by:", addr, "titled:", req.PostFormValue("title"))

	return nil
}

// AfterAcceptUpdate is called after AcceptUpdate, and is useful for logging or triggering
// notifications, etc. after the data is saved to the database, etc.
// The request has a context containing the databse 'target' affected by the
// request.
func (s *Song) AfterAcceptUpdate(res http.ResponseWriter, req *http.Request) error {
	addr := req.RemoteAddr
	log.Println("Song updated by:", addr, "titled:", req.PostFormValue("title"))

	return nil
}

// Approve implements editor.Mergeable, which enables content supplied by external
// clients to be approved and thus added to the public content API. Before content
// is approved, it is waiting in the Pending bucket, and can only be approved in
// the CMS if the Mergeable interface is satisfied. If not, you will not see this
// content show up in the CMS.
func (s *Song) Approve(res http.ResponseWriter, req *http.Request) error {
	return nil
}

/*
   NOTICE: if AutoApprove (seen below) is implemented, the Approve method above will have no
   effect, except to add the Public / Pending toggle in the CMS UI. Though, no
   Song content would be in Pending, since all externally submitting Song data
   is immediately approved.
*/

// AutoApprove implements api.Trustable, and will automatically approve content
// that has been submitted by an external client via api.Externalable. Be careful
// when using AutoApprove, because content will immediately be available through
// your public content API. If the Trustable interface is satisfied, the AfterApprove
// method is bypassed. The
func (s *Song) AutoApprove(res http.ResponseWriter, req *http.Request) error {
	// Use AutoApprove to check for trust-specific headers or whitelisted IPs,
	// etc. Remember, you will not be able to Approve or Reject content that
	// is auto-approved. You could add a field to Song, i.e.
	// AutoApproved bool `json:auto_approved`
	// and set that data here, as it is called before the content is saved, but
	// after the BeforeSave hook.

	return nil
}
