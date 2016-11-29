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
	TypeName string
	JSONName string
}

// blog title:string Author:string PostCategory:string content:string some_thing:int
func parseType(args []string) (generateType, error) {
	t := generateType{
		Name: fieldName(args[0]),
	}
	t.Initial = strings.ToLower(string(t.Name[0]))

	fields := args[1:]
	for _, field := range fields {
		f, err := parseField(field)
		if err != nil {
			return generateType{}, err
		}

		t.Fields = append(t.Fields, f)
	}

	return t, nil
}

func parseField(raw string) (generateField, error) {
	// title:string
	if !strings.Contains(raw, ":") {
		return generateField{}, fmt.Errorf("Invalid generate argument. [%s]", raw)
	}

	pair := strings.Split(raw, ":")
	field := generateField{
		Name:     fieldName(pair[0]),
		TypeName: strings.ToLower(pair[1]),
		JSONName: fieldJSONName(pair[0]),
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

	tmplPath := filepath.Join(pwd, "cmd", "ponzu", "contentType.tmpl")
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
