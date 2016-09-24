package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
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

func newProjectInDir(path string) error {
	// check if anything exists at the path, ask if it should be overwritten
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		fmt.Println("Path exists, overwrite contents? (y/N):")
		// input := bufio.NewReader(os.Stdin)
		// answer, err := input.ReadString('\n')

		var answer string
		_, err := fmt.Scanf("%s\n", &answer)
		if err != nil {
			return err
		}

		answer = strings.ToLower(answer)

		switch answer {
		case "n", "no", "":
			fmt.Println("")

		case "y", "yes":
			err := os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("Failed to overwrite %s. \n%s", path, err)
			}

			return createProjInDir(path)

		default:
			fmt.Println("Input not recognized. No files overwritten. Answer as 'y' or 'n' only.")
		}

		return nil
	}

	return createProjInDir(path)
}

func createProjInDir(path string) error {
	var buf = &bytes.Buffer{}
	echo := exec.Command("echo", os.Getenv("GOPATH"))
	echo.Stdout = buf
	err := echo.Run()
	if err != nil {
		return err
	}

	gopath := buf.String()
	gopath = gopath[:len(gopath)-1]
	gopath = filepath.Join(gopath, "src")

	repo := "github.com/nilslice/cms"
	local := filepath.Join(gopath, repo)
	network := "https://" + repo + ".git"

	// create the directory or overwrite it
	err = os.MkdirAll(path, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	// try to git clone the repository from the local machine's $GOPATH
	localClone := exec.Command("git", "clone", local, path)
	localClone.Stdout = os.Stdout
	localClone.Stderr = os.Stderr

	err = localClone.Start()
	if err != nil {
		return err
	}
	err = localClone.Wait()
	if err != nil {
		fmt.Println("Couldn't clone from", local, ". Trying network...")

		// try to git clone the repository over the network
		networkClone := exec.Command("git", "clone", network, path)
		networkClone.Stdout = os.Stdout
		networkClone.Stderr = os.Stderr

		err = networkClone.Start()
		if err != nil {
			fmt.Println("Network clone failed to start. Try again and make sure you have a network connection.")
			return err
		}
		err = networkClone.Wait()
		if err != nil {
			fmt.Println("Network clone failure.")
			// failed
			return fmt.Errorf("Failed to clone files from local machine [%s] and over the network [%s].\n%s", local, network, err)
		}
	}

	fmt.Println("New project created at", path)
	return nil
}
