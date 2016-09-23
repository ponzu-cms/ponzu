package content

import (
	"fmt"

	"github.com/nilslice/cms/management/editor"
)

// Post is the generic content struct
type Post struct {
	Item
	editor editor.Editor

	Title     string `json:"title"`
	Content   string `json:"content"`
	Author    string `json:"author"`
	Timestamp string `json:"timestamp"`
}

func init() {
	Types["Post"] = func() interface{} { return new(Post) }
}

// SetContentID partially implements editor.Editable
func (p *Post) SetContentID(id int) { p.ID = id }

// ContentID partially implements editor.Editable
func (p *Post) ContentID() int { return p.ID }

// ContentName partially implements editor.Editable
func (p *Post) ContentName() string { return p.Title }

// SetSlug partially implements editor.Editable
func (p *Post) SetSlug(slug string) { p.Slug = slug }

// Editor partially implements editor.Editable
func (p *Post) Editor() *editor.Editor { return &p.editor }

// MarshalEditor writes a buffer of html to edit a Post and partially implements editor.Editable
func (p *Post) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(p,
		editor.Field{
			View: editor.Input("Slug", p, map[string]string{
				"label":       "URL Path",
				"type":        "text",
				"disabled":    "true",
				"placeholder": "Will be set automatically",
			}),
		},
		editor.Field{
			View: editor.Input("Title", p, map[string]string{
				"label":       "Post Title",
				"type":        "text",
				"placeholder": "Enter your Post Title here",
			}),
		},
		editor.Field{
			View: editor.Textarea("Content", p, map[string]string{
				"label":       "Content",
				"placeholder": "Add the content of your post here",
			}),
		},
		editor.Field{
			View: editor.Input("Author", p, map[string]string{
				"label":       "Author",
				"type":        "text",
				"placeholder": "Enter the author name here",
			}),
		},
		editor.Field{
			View: editor.Input("Timestamp", p, map[string]string{
				"label": "Publish Date",
				"type":  "date",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Post editor view: %s", err.Error())
	}

	return view, nil
}
