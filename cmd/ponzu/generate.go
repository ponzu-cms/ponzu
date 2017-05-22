package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

type generateType struct {
	Name          string
	Initial       string
	Fields        []generateField
	HasReferences bool
}

type generateField struct {
	Name     string
	Initial  string
	TypeName string
	JSONName string
	View     string

	IsReference       bool
	ReferenceName     string
	ReferenceJSONTags []string
}

var reservedFieldNames = map[string]string{
	"uuid":      "UUID",
	"item":      "Item",
	"id":        "ID",
	"slug":      "Slug",
	"timestamp": "Timestamp",
	"updated":   "Updated",
}

func legalFieldNames(fields ...generateField) (bool, map[string]string) {
	conflicts := make(map[string]string)
	for _, field := range fields {
		for jsonName, fieldName := range reservedFieldNames {
			if field.JSONName == jsonName || field.Name == fieldName {
				conflicts[jsonName] = fieldName
			}
		}
	}

	if len(conflicts) > 0 {
		return false, conflicts
	}

	return true, conflicts
}

// blog title:string Author:string PostCategory:string content:string some_thing:int
func parseType(args []string) (generateType, error) {
	t := generateType{
		Name: fieldName(args[0]),
	}
	t.Initial = strings.ToLower(string(t.Name[0]))

	fields := args[1:]
	for _, field := range fields {
		f, err := parseField(field, &t)
		if err != nil {
			return generateType{}, err
		}

		// set initial (1st character of the type's name) on field so we don't need
		// to set the template variable like was done in prior version
		f.Initial = t.Initial

		t.Fields = append(t.Fields, f)
	}

	if ok, nameConflicts := legalFieldNames(t.Fields...); !ok {
		for jsonName, fieldName := range nameConflicts {
			fmt.Println(fmt.Sprintf("reserved field name: %s (%s)", jsonName, fieldName))
		}

		count := len(nameConflicts)
		var c = "conflict"
		if count > 1 {
			c = "conflicts"
		}

		return generateType{}, fmt.Errorf("You have (%d) naming %s - please rename and try again", count, c)
	}

	return t, nil
}

func parseField(raw string, gt *generateType) (generateField, error) {
	// contents:string
	// contents:string:richtext
	// author:@author,name,age
	// authors:[]@author,name,age

	if !strings.Contains(raw, ":") {
		return generateField{}, fmt.Errorf("Invalid generate argument. [%s]", raw)
	}

	data := strings.Split(raw, ":")

	field := generateField{
		Name:     fieldName(data[0]),
		Initial:  gt.Initial,
		JSONName: fieldJSONName(data[0]),
	}

	setFieldTypeName(&field, data[1], gt)

	viewType := "input"
	if len(data) == 3 {
		viewType = data[2]
	}

	err := setFieldView(&field, viewType)
	if err != nil {
		return generateField{}, err
	}

	return field, nil
}

// parse the field's type name and check if it is a special reference type, or
// a slice of reference types, which we'll set their underlying type to string
// or []string respectively
func setFieldTypeName(field *generateField, fieldType string, gt *generateType) {
	if !strings.Contains(fieldType, "@") {
		// not a reference, set as-is downcased
		field.TypeName = strings.ToLower(fieldType)
		field.IsReference = false
		return
	}

	// some possibilities are
	// @author,name,age
	// []@author,name,age
	// -------------------
	// [] = slice of author
	// @author = reference to Author struct
	// ,name,age = JSON tag names from Author struct fields to use as select option display

	if strings.Contains(fieldType, ",") {
		referenceConf := strings.Split(fieldType, ",")
		fieldType = referenceConf[0]
		field.ReferenceJSONTags = referenceConf[1:]
	}

	var referenceType string
	if strings.HasPrefix(fieldType, "[]") {
		referenceType = strings.TrimPrefix(fieldType, "[]@")
		fieldType = "[]string"
	} else {
		referenceType = strings.TrimPrefix(fieldType, "@")
		fieldType = "string"
	}

	field.TypeName = strings.ToLower(fieldType)
	field.ReferenceName = fieldName(referenceType)
	field.IsReference = true
	gt.HasReferences = true
	return
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

func optimizeFieldView(field *generateField, viewType string) string {
	viewType = strings.ToLower(viewType)

	if field.IsReference {
		viewType = "reference"
	}

	// if we have a []T field type, automatically make the input view a repeater
	// as long as a repeater exists for the input type
	repeaterElements := []string{"input", "select", "file", "reference"}
	if strings.HasPrefix(field.TypeName, "[]") {
		for _, el := range repeaterElements {
			// if the viewType already is declared to be a -repeater
			// the comparison below will fail but the switch will
			// still find the right generator template
			// ex. authors:"[]string":select
			// ex. authors:string:select-repeater
			if viewType == el {
				viewType = viewType + "-repeater"
			}
		}
	} else {
		// if the viewType is already declared as a -repeater, but
		// the TypeName is not of []T, add the [] prefix so the user
		// code is correct
		// ex. authors:string:select-repeater
		// ex. authors:@author:select-repeater
		if strings.HasSuffix(viewType, "-repeater") {
			field.TypeName = "[]" + field.TypeName
		}
	}

	return viewType
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
	tmplFromWithDelims := func(filename string, delim [2]string) (*template.Template, error) {
		if delim[0] == "" || delim[1] == "" {
			delim = [2]string{"{{", "}}"}
		}

		return template.New(filename).Delims(delim[0], delim[1]).ParseFiles(filepath.Join(tmplDir, filename))
	}

	viewType = optimizeFieldView(field, viewType)
	switch viewType {
	case "checkbox":
		tmpl, err = tmplFromWithDelims("gen-checkbox.tmpl", [2]string{})
	case "custom":
		tmpl, err = tmplFromWithDelims("gen-custom.tmpl", [2]string{})
	case "file":
		tmpl, err = tmplFromWithDelims("gen-file.tmpl", [2]string{})
	case "hidden":
		tmpl, err = tmplFromWithDelims("gen-hidden.tmpl", [2]string{})
	case "input", "text":
		tmpl, err = tmplFromWithDelims("gen-input.tmpl", [2]string{})
	case "richtext":
		tmpl, err = tmplFromWithDelims("gen-richtext.tmpl", [2]string{})
	case "select":
		tmpl, err = tmplFromWithDelims("gen-select.tmpl", [2]string{})
	case "textarea":
		tmpl, err = tmplFromWithDelims("gen-textarea.tmpl", [2]string{})
	case "tags":
		tmpl, err = tmplFromWithDelims("gen-tags.tmpl", [2]string{})

	case "input-repeater":
		tmpl, err = tmplFromWithDelims("gen-input-repeater.tmpl", [2]string{})
	case "select-repeater":
		tmpl, err = tmplFromWithDelims("gen-select-repeater.tmpl", [2]string{})
	case "file-repeater":
		tmpl, err = tmplFromWithDelims("gen-file-repeater.tmpl", [2]string{})

	// use [[ and ]] as delimeters since reference views need to generate
	// display names containing {{ and }}
	case "reference":
		tmpl, err = tmplFromWithDelims("gen-reference.tmpl", [2]string{"[[", "]]"})
		if err != nil {
			return err
		}
	case "reference-repeater":
		tmpl, err = tmplFromWithDelims("gen-reference-repeater.tmpl", [2]string{"[[", "]]"})
		if err != nil {
			return err
		}

	default:
		msg := fmt.Sprintf("'%s' is not a recognized view type. Using 'input' instead.", viewType)
		fmt.Println(msg)
		tmpl, err = tmplFromWithDelims("gen-input.tmpl", [2]string{})
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
		localFile := filepath.Join("content", fileName)
		return fmt.Errorf("Please remove '%s' before executing this command", localFile)
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

	// no file exists.. ok to write new one
	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		return err
	}

	_, err = file.Write(fmtBuf)
	if err != nil {
		return fmt.Errorf("Failed to write generated file buffer: %s", err.Error())
	}

	return nil
}

var generateCmd = &cobra.Command{
	Use:     "generate <generator type (,...fields)>",
	Aliases: []string{"gen", "g"},
	Short:   "generate boilerplate code for various Ponzu components",
	Long: `Generate boilerplate code for various Ponzu components, such as 'content'.

The command above will generate a file 'content/review.go' with boilerplate
methods, as well as struct definition, and corresponding field tags like:

type Review struct {
	Title  string   ` + "`json:" + `"title"` + "`" + `
	Body   string   ` + "`json:" + `"body"` + "`" + `
	Rating int      ` + "`json:" + `"rating"` + "`" + `
	Tags   []string ` + "`json:" + `"tags"` + "`" + `
}

The generate command will intelligently parse more sophisticated field names
such as 'field_name' and convert it to 'FieldName' and vice versa, only where
appropriate as per common Go idioms. Errors will be reported, but successful
generate commands return nothing.`,
	Example: `$ ponzu gen content review title:"string" body:"string" rating:"int" tags:"[]string"`,
}

var contentCmd = &cobra.Command{
	Use:     "content <namespace> <field> <field>...",
	Aliases: []string{"c"},
	Short:   "generates a new content type",
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateContentType(args)
	},
}

func init() {
	generateCmd.AddCommand(contentCmd)
	RegisterCmdlineCommand(generateCmd)
}
