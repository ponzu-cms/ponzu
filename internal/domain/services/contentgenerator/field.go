package contentgenerator

import (
	"bytes"
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"path/filepath"
	"strings"

	"text/template"
)

var (
	reservedFieldNames = map[string]string{
		"uuid":      "UUID",
		"item":      "Item",
		"id":        "ID",
		"slug":      "Slug",
		"timestamp": "Timestamp",
		"updated":   "Updated",
	}
)

func (gt *generator) ValidateField(field *entities.Field) error {
	for jsonName, fieldName := range reservedFieldNames {
		if field.JSONName == jsonName || field.Name == fieldName {
			return fmt.Errorf("reserved field name: %s (%s)", jsonName, fieldName)
		}
	}

	return nil
}

// set the specified view inside the editor field for a generated field for a type
func (gt *generator) setFieldView(field *entities.Field) error {
	var err error
	var tmpl *template.Template
	buf := &bytes.Buffer{}

	tmplFromWithDelims := func(filename string, delim [2]string) (*template.Template, error) {
		if delim[0] == "" || delim[1] == "" {
			delim = [2]string{"{{", "}}"}
		}

		return template.New(filename).Delims(delim[0], delim[1]).ParseFiles(filepath.Join(gt.templateDir, filename))
	}

	optimizeFieldView(field)
	switch field.ViewType {
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

	// use [[ and ]] as delimiters since reference views need to generate
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
		msg := fmt.Sprintf("'%s' is not a recognized view type. Using 'input' instead.", field.ViewType)
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

func optimizeFieldView(field *entities.Field) {
	field.ViewType = strings.ToLower(field.ViewType)

	if field.IsReference {
		field.ViewType = "reference"
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
			if field.ViewType == el {
				field.ViewType = field.ViewType + "-repeater"
			}
		}
	} else {
		// if the viewType is already declared as a -repeater, but
		// the TypeName is not of []T, add the [] prefix so the user
		// code is correct
		// ex. authors:string:select-repeater
		// ex. authors:@author:select-repeater
		if strings.HasSuffix(field.ViewType, "-repeater") {
			field.TypeName = "[]" + field.TypeName
		}
	}
}
