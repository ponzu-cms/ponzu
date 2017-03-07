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
	"strings"

	"github.com/ponzu-cms/ponzu/management/editor"
	"github.com/ponzu-cms/ponzu/system/addon"
)

// Select returns the []byte of a <select> HTML element plus internal <options> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Select(fieldName string, p interface{}, attrs map[string]string, contentType, tmplString string) []byte {
	options, err := encodeDataToOptions(contentType, tmplString)
	if err != nil {
		log.Println("Error encoding data to options for", contentType, err)
		return nil
	}

	return editor.Select(fieldName, p, attrs, options)
}

// SelectRepeater returns the []byte of a <select> HTML element plus internal <options> with a label.
// It also includes repeat controllers (+ / -) so the element can be
// dynamically multiplied or reduced.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func SelectRepeater(fieldName string, p interface{}, attrs map[string]string, contentType, tmplString string) []byte {
	scope := editor.TagNameFromStructField(fieldName, p)
	html := bytes.Buffer{}
	_, err := html.WriteString(`<span class="__ponzu-repeat ` + scope + `">`)
	if err != nil {
		log.Println("Error writing HTML string to SelectRepeater buffer")
		return nil
	}

	if _, ok := attrs["class"]; ok {
		attrs["class"] += " browser-default"
	} else {
		attrs["class"] = "browser-default"
	}

	// find the field values in p to determine if an option is pre-selected
	fieldVals := editor.ValueFromStructField(fieldName, p)
	vals := strings.Split(fieldVals, "__ponzu")

	options, err := encodeDataToOptions(contentType, tmplString)
	if err != nil {
		log.Println("Error encoding data to options for", contentType, err)
		return nil
	}

	for _, val := range vals {
		sel := editor.NewElement("select", attrs["label"], fieldName, p, attrs)
		var opts []*editor.Element

		// provide a call to action for the select element
		cta := &editor.Element{
			TagName: "option",
			Attrs:   map[string]string{"disabled": "true", "selected": "true"},
			Data:    "Select an option...",
			ViewBuf: &bytes.Buffer{},
		}

		// provide a selection reset (will store empty string in db)
		reset := &editor.Element{
			TagName: "option",
			Attrs:   map[string]string{"value": ""},
			Data:    "None",
			ViewBuf: &bytes.Buffer{},
		}

		opts = append(opts, cta, reset)

		for k, v := range options {
			optAttrs := map[string]string{"value": k}
			if k == val {
				optAttrs["selected"] = "true"
			}
			opt := &editor.Element{
				TagName: "option",
				Attrs:   optAttrs,
				Data:    v,
				ViewBuf: &bytes.Buffer{},
			}

			opts = append(opts, opt)
		}

		_, err := html.Write(editor.DOMElementWithChildrenSelect(sel, opts))
		if err != nil {
			log.Println("Error writing DOMElementWithChildrenSelect to SelectRepeater buffer")
			return nil
		}
	}

	_, err = html.WriteString("</span>")
	if err != nil {
		log.Println("Error writing HTML string to SelectRepeater buffer")
		return nil
	}

	return append(html.Bytes(), editor.RepeatController(fieldName, p, "select", ".input-field")...)
}

func encodeDataToOptions(contentType, tmplString string) (map[string]string, error) {
	// encode all content type from db into options map
	// options in form of map["/api/content?type=<contentType>&id=<id>"]t.String()
	options := make(map[string]string)

	var all map[string]interface{}
	j := addon.ContentAll(contentType)

	err := json.Unmarshal(j, &all)
	if err != nil {
		return nil, err
	}

	// make template for option html display
	tmpl := template.Must(template.New(contentType).Parse(tmplString))

	// make data something usable to iterate over and assign options
	data := all["data"].([]interface{})

	for i := range data {
		item := data[i].(map[string]interface{})
		k := fmt.Sprintf("/api/content?type=%s&id=%.0f", contentType, item["id"].(float64))
		v := &bytes.Buffer{}
		err := tmpl.Execute(v, item)
		if err != nil {
			return nil, fmt.Errorf(
				"Error executing template for reference of %s: %s",
				contentType, err.Error())
		}

		options[k] = v.String()
	}

	return options, nil
}
