package content

import (
	"fmt"

	"github.com/bosssauce/ponzu/management/editor"
)

// Post is the generic content struct
type Post struct {
	Item
	editor editor.Editor

	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Photo      string   `json:"photo"`
	Author     string   `json:"author"`
	Category   []string `json:"category"`
	ThemeStyle string   `json:"theme"`
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
			View: editor.Input("Title", p, map[string]string{
				"label":       "Post Title",
				"type":        "text",
				"placeholder": "Enter your Post Title here",
			}),
		},
		editor.Field{
			View: editor.Richtext("Content", p, map[string]string{
				"label":       "Content",
				"placeholder": "Add the content of your post here",
			}),
		},
		editor.Field{
			View: editor.File("Picture", p, map[string]string{
				"label":       "Author Photo",
				"placeholder": "Upload a profile picture for the author",
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
			View: editor.Checkbox("Category", p, map[string]string{
				"label": "Post Category",
			}, map[string]string{
				"important": "Important",
				"active":    "Active",
				"unplanned": "Unplanned",
			}),
		},
		editor.Field{
			View: editor.Select("ThemeStyle", p, map[string]string{
				"label": "Theme Style",
			}, map[string]string{
				"dark":  "Dark",
				"light": "Light",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Post editor view: %s", err.Error())
	}

	return view, nil
}
