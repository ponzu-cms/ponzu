// Package editor enables users to create edit views from their content
// structs so that admins can manage content
package editor

import "bytes"

// Editable ensures data is editable
type Editable interface {
	ContentID() int
	ContentName() string
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

// New takes editable content and any number of Field funcs to describe the edit
// page for any content struct added by a user
func New(post Editable, fields ...Field) ([]byte, error) {
	editor := post.Editor()

	editor.ViewBuf = &bytes.Buffer{}

	for _, f := range fields {
		addFieldToEditorView(editor, f)
	}

	return editor.ViewBuf.Bytes(), nil
}

func addFieldToEditorView(e *Editor, f Field) {
	e.ViewBuf.Write(f.View)
}
