package main

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/nilslice/cms/editor"
)

// Post is the generic content struct
type Post struct {
	editor editor.Editor

	Title     []byte `json:"title"`
	Content   []byte `json:"content"`
	Author    []byte `json:"author"`
	Timestamp []byte `json:"timestamp"`
}

// Editor partially implements editor.Editable
func (p *Post) Editor() *editor.Editor {
	return &p.editor
}

// NewViewBuffer partially implements editor.Editable
func (p *Post) NewViewBuffer() {
	p.editor.ViewBuf = &bytes.Buffer{}
}

// Render partially implements editor.Editable
func (p *Post) Render() []byte {
	return p.editor.ViewBuf.Bytes()
}

// EditView writes a buffer of html to edit a Post
func (p Post) EditView() ([]byte, error) {
	view, err := editor.New(&p,
		editor.Field{
			View: editor.Input("Title", &p, map[string]string{
				"label":       "Post Title",
				"type":        "text",
				"placeholder": "Enter your Post Title here",
			}),
		},
		editor.Field{
			View: editor.Textarea("Content", &p, map[string]string{
				"label":       "Content",
				"placeholder": "Add the content of your post here",
			}),
		},
		editor.Field{
			View: editor.Input("Author", &p, map[string]string{
				"label":       "Author",
				"type":        "text",
				"placeholder": "Enter the author name here",
			}),
		},
		editor.Field{
			View: editor.Input("Timestamp", &p, map[string]string{
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

func (p Post) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-type", "text/html")
	resp, err := p.EditView()
	if err != nil {
		fmt.Println(err)
	}
	res.Write(resp)
}

func main() {
	p := Post{
		Content:   []byte("<h3>H</h3>ello. My name is <em>Steve</em>."),
		Title:     []byte("Profound introduction"),
		Author:    []byte("Steve Manuel"),
		Timestamp: []byte("2016-09-16"),
	}

	http.ListenAndServe(":8080", p)

}
