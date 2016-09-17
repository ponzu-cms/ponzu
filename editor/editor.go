// Package editor enables users to create edit views from their content
// structs so that admins can manage content
package editor

import "bytes"

// Editable ensures data is editable
type Editable interface {
	Editor() *Editor
	NewViewBuffer()
	Render() []byte
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

// New takes editable content and any number of Field funcs to describe the edit
// page for any content struct added by a user
func New(post Editable, fields ...Field) ([]byte, error) {
	post.NewViewBuffer()

	editor := post.Editor()

	for _, f := range fields {
		addFieldToEditorView(editor, f)
	}

	return post.Render(), nil
}

func addFieldToEditorView(e *Editor, f Field) {
	e.ViewBuf.Write(f.View)
}
