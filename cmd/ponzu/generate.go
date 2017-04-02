package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type generateType struct {
	Name    string
	Initial string
	Fields  []generateField
}

type generateField struct {
	Name     string
	Initial  string
	TypeName string
	JSONName string
	View     string
}

// blog title:string Author:string PostCategory:string content:string some_thing:int
func parseType(args []string) (generateType, error) {
	t := generateType{
		Name: fieldName(args[0]),
	}
	t.Initial = strings.ToLower(string(t.Name[0]))

	fields := args[1:]
	for _, field := range fields {
		f, err := parseField(field, t)
		if err != nil {
			return generateType{}, err
		}
		// NEW
		// set initial (1st character of the type's name) on field so we don't need
		// to set the template variable like was done in prior version
		f.Initial = t.Initial

		t.Fields = append(t.Fields, f)
	}

	return t, nil
}

func parseField(raw string, gt generateType) (generateField, error) {
	// contents:string or // contents:string:richtext
	if !strings.Contains(raw, ":") {
		return generateField{}, fmt.Errorf("Invalid generate argument. [%s]", raw)
	}

	data := strings.Split(raw, ":")

	field := generateField{
		Name:     fieldName(data[0]),
		Initial:  gt.Initial,
		TypeName: strings.ToLower(data[1]),
		JSONName: fieldJSONName(data[0]),
	}

	fieldType := "input"
	if len(data) == 3 {
		fieldType = data[2]
	}

	err := setFieldView(&field, fieldType)
	if err != nil {
		return generateField{}, err
	}

	return field, nil
}

// get the initial field name passed and check it for all possible cases
// MyTitle:string myTitle:string my_title:string -> MyTitle
// error-message:string -> ErrorMessage
func fieldName(name string) string {
	// remove _ or - if first character
	if name[0] == '-' || name[0] == '_' {
		name = name[1:]
	}

	// remove _ or - if last character
	if name[len(name)-1] == '-' || name[len(name)-1] == '_' {
		name = name[:len(name)-1]
	}

	// upcase the first character
	name = strings.ToUpper(string(name[0])) + name[1:]

	// remove _ or - character, and upcase the character immediately following
	for i := 0; i < len(name); i++ {
		r := rune(name[i])
		if isUnderscore(r) || isHyphen(r) {
			up := strings.ToUpper(string(name[i+1]))
			name = name[:i] + up + name[i+2:]
		}
	}

	return name
}

// get the initial field name passed and convert to json-like name
// MyTitle:string myTitle:string my_title:string -> my_title
// error-message:string -> error-message
func fieldJSONName(name string) string {
	// remove _ or - if first character
	if name[0] == '-' || name[0] == '_' {
		name = name[1:]
	}

	// downcase the first character
	name = strings.ToLower(string(name[0])) + name[1:]

	// check for uppercase character, downcase and insert _ before it if i-1
	// isn't already _ or -
	for i := 0; i < len(name); i++ {
		r := rune(name[i])
		if isUpper(r) {
			low := strings.ToLower(string(r))
			if name[i-1] == '_' || name[i-1] == '-' {
				name = name[:i] + low + name[i+1:]
			} else {
				name = name[:i] + "_" + low + name[i+1:]
			}
		}
	}

	return name
}

// set the specified view inside the editor field for a generated field for a type
func setFieldView(field *generateField, viewType string) error {
	var err error
	var tmpl *template.Template
	buf := &bytes.Buffer{}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	tmplDir := filepath.Join(pwd, "cmd", "ponzu", "templates")
	tmplFrom := func(filename string) (*template.Template, error) {
		return template.ParseFiles(filepath.Join(tmplDir, filename))
	}

	viewType = strings.ToLower(viewType)
	switch viewType {
	case "checkbox":
		tmpl, err = tmplFrom("gen-checkbox.tmpl")
	case "custom":
		tmpl, err = tmplFrom("gen-custom.tmpl")
	case "file":
		tmpl, err = tmplFrom("gen-file.tmpl")
	case "hidden":
		tmpl, err = tmplFrom("gen-hidden.tmpl")
	case "input", "text":
		tmpl, err = tmplFrom("gen-input.tmpl")
	case "richtext":
		tmpl, err = tmplFrom("gen-richtext.tmpl")
	case "select":
		tmpl, err = tmplFrom("gen-select.tmpl")
	case "textarea":
		tmpl, err = tmplFrom("gen-textarea.tmpl")
	case "tags":
		tmpl, err = tmplFrom("gen-tags.tmpl")
	default:
		msg := fmt.Sprintf("'%s' is not a recognized view type. Using 'input' instead.", viewType)
		fmt.Println(msg)
		tmpl, err = tmplFrom("gen-input.tmpl")
	}

	if err != nil {
		return err
	}

	err = tmpl.Execute(buf, field)
	if err != nil {
		return err
	}

	field.View = buf.String()

	return nil
}

func isUpper(char rune) bool {
	if char >= 'A' && char <= 'Z' {
		return true
	}

	return false
}

func isUnderscore(char rune) bool {
	return char == '_'
}

func isHyphen(char rune) bool {
	return char == '-'
}

func generateContentType(args []string) error {
	name := args[0]
	fileName := strings.ToLower(name) + ".go"

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

	// parse type info from args
	gt, err := parseType(args)
	if err != nil {
		return fmt.Errorf("Failed to parse type args: %s", err.Error())
	}

	tmplPath := filepath.Join(pwd, "cmd", "ponzu", "templates", "gen-content.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("Failed to parse template: %s", err.Error())
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, gt)
	if err != nil {
		return fmt.Errorf("Failed to execute template: %s", err.Error())
	}

	fmtBuf, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("Failed to format template: %s", err.Error())
	}

	_, err = file.Write(fmtBuf)
	if err != nil {
		return fmt.Errorf("Failed to write generated file buffer: %s", err.Error())
	}

	return nil
}
