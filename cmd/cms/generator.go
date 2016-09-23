package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

func generateContentType(name string) error {
	fileName := strings.ToLower(name) + ".go"
	typeName := strings.ToUpper(string(name[0])) + string(name[1:])

	// contain processed name an info for template
	data := map[string]string{
		"name":    typeName,
		"initial": string(fileName[0]),
	}

	// open file in ./content/ dir
	// if exists, alert user of conflict
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	contentDir := filepath.Join(pwd, "content")
	filePath := filepath.Join(contentDir, fileName)

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return fmt.Errorf("Please remove '%s' before executing this command.", fileName)
	}

	// no file exists.. ok to write new one
	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		return err
	}

	// execute template
	tmpl := template.Must(template.New("content").Parse(contentTypeTmpl))
	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}

const contentTypeTmpl = `
package content

import (
	"fmt"

	"github.com/nilslice/cms/management/editor"
)

// {{ .name }} is the generic content struct
type {{ .name }} struct {
	Item
	editor editor.Editor

/*  
    // all your custom fields must have json tags!
	Title     string ` + "`json:" + `"title"` + "`" + `
	Content   string ` + "`json:" + `"content"` + "`" + `
	Author    string ` + "`json:" + `"author"` + "`" + `
	Timestamp string ` + "`json:" + `"timestamp"` + "`" + `
*/
}

func init() {
	Types["{{ .name }}"] = func() interface{} { return new({{ .name }}) }
}

// SetContentID partially implements editor.Editable
func ({{ .initial }} *{{ .name }}) SetContentID(id int) { {{ .initial }}.ID = id }

// ContentID partially implements editor.Editable
func ({{ .initial }} *{{ .name }}) ContentID() int { return {{ .initial }}.ID }

// ContentName partially implements editor.Editable
func ({{ .initial }} *{{ .name }}) ContentName() string { return {{ .initial }}.Title }

// SetSlug partially implements editor.Editable
func ({{ .initial }} *{{ .name }}) SetSlug(slug string) { {{ .initial }}.Slug = slug }

// Editor partially implements editor.Editable
func ({{ .initial }} *{{ .name }}) Editor() *editor.Editor { return &{{ .initial }}.editor }

// MarshalEditor writes a buffer of html to edit a {{ .name }} and partially implements editor.Editable
func ({{ .initial }} *{{ .name }}) MarshalEditor() ([]byte, error) {
/* EXAMPLE CODE (from post.go, the default content type)
	view, err := editor.Form({{ .initial }},
		editor.Field{
            // Take careful note that the first argument to these Input-like methods 
            // is the string version of each {{ .name }} struct tag, and must follow this pattern
            // for auto-decoding and -encoding reasons.
			View: editor.Input("Slug", {{ .initial }}, map[string]string{
				"label":       "URL Path",
				"type":        "text",
				"disabled":    "true",
				"placeholder": "Will be set automatically",
			}),
		},
		editor.Field{
			View: editor.Input("Title", {{ .initial }}, map[string]string{
				"label":       "{{ .name }} Title",
				"type":        "text",
				"placeholder": "Enter your {{ .name }} Title here",
			}),
		},
		editor.Field{
			View: editor.Textarea("Content", {{ .initial }}, map[string]string{
				"label":       "Content",
				"placeholder": "Add the content of your {{ .name }} here",
			}),
		},
		editor.Field{
			View: editor.Input("Author", {{ .initial }}, map[string]string{
				"label":       "Author",
				"type":        "text",
				"placeholder": "Enter the author name here",
			}),
		},
		editor.Field{
			View: editor.Input("Timestamp", {{ .initial }}, map[string]string{
				"label": "Publish Date",
				"type":  "date",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render {{ .name }} editor view: %s", err.Error())
	}

	return view, nil
*/
}
`
