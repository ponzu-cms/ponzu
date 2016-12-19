// Package reference is a Ponzu addon to enable content editors to create
// references to other content types which are stored as query strings within
// the referencer's content DB
package reference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"

	"github.com/ponzu-cms/ponzu/management/editor"
	"github.com/ponzu-cms/ponzu/system/addon"
)

// Select returns the []byte of a <select> HTML element plus internal <options> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Select(fieldName string, p interface{}, attrs map[string]string, contentType, tmplString string) []byte {
	// decode all content type from db into options map
	// options in form of map["?type=<contentType>&id=<id>"]t.String()
	options := make(map[string]string)

	var all map[string]interface{}
	j := addon.ContentAll(contentType)

	err := json.Unmarshal(j, &all)
	if err != nil {
		return nil
	}

	// make template for option html display
	tmpl := template.Must(template.New(contentType).Parse(tmplString))

	// make data something usable to iterate over and assign options
	data := all["data"].([]interface{})

	for i := range data {
		item := data[i].(map[string]interface{})
		k := fmt.Sprintf("?type=%s&id=%.0f", contentType, item["id"].(float64))
		v := &bytes.Buffer{}
		err := tmpl.Execute(v, item)
		if err != nil {
			log.Println("Error executing template for reference of:", contentType)
			return nil
		}

		options[k] = v.String()
	}

	return editor.Select(fieldName, p, attrs, options)
}
