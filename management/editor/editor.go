// Package editor enables users to create edit views from their content
// structs so that admins can manage content
package editor

import (
	"bytes"
)

// Editable ensures data is editable
type Editable interface {
	SetContentID(id int)
	ContentID() int
	ContentName() string
	SetSlug(slug string)
	Editor() *Editor
	MarshalEditor() ([]byte, error)
}

// Editor is a view containing fields to manage content
type Editor struct {
	ViewBuf *bytes.Buffer
}

// Field is used to create the editable view for a field
// within a particular content struct
type Field struct {
	View []byte
}

// Form takes editable content and any number of Field funcs to describe the edit
// page for any content struct added by a user
func Form(post Editable, fields ...Field) ([]byte, error) {
	editor := post.Editor()

	editor.ViewBuf = &bytes.Buffer{}
	editor.ViewBuf.Write([]byte(`<table><tbody class="row"><tr class="col s8"><td>`))

	for _, f := range fields {
		addFieldToEditorView(editor, f)
	}

	editor.ViewBuf.Write([]byte(`</td></tr>`))

	// content items with Item embedded have some default fields we need to render
	editor.ViewBuf.Write([]byte(`<tr class="col s4"><td>`))
	addPostDefaultFieldsToEditorView(post, editor)

	submit := `
<div class="input-field">
	<input class="right waves-effect waves-light btn green" type="submit" value="Save"/>
</div>
`
	editor.ViewBuf.Write([]byte(submit + `</td></tr></tbody></table>`))

	return editor.ViewBuf.Bytes(), nil
}

func addFieldToEditorView(e *Editor, f Field) {
	e.ViewBuf.Write(f.View)
}

func addPostDefaultFieldsToEditorView(p Editable, e *Editor) {
	defaults := []Field{
		Field{
			View: Input("Timestamp", p, map[string]string{
				"label": "Publish Date",
				"type":  "date",
			}),
		},
		Field{
			View: Input("Slug", p, map[string]string{
				"label":       "URL Slug",
				"type":        "text",
				"disabled":    "true",
				"placeholder": "Will be set automatically",
			}),
		},
	}

	for _, f := range defaults {
		addFieldToEditorView(e, f)
	}

}
