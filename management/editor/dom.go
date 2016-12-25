package editor

import (
	"bytes"
	"html"
	"strings"
)

type element struct {
	tagName string
	attrs   map[string]string
	name    string
	label   string
	data    string
	viewBuf *bytes.Buffer
}

func newElement(tagName, label, fieldName string, p interface{}, attrs map[string]string) *element {
	return &element{
		tagName: tagName,
		attrs:   attrs,
		name:    tagNameFromStructField(fieldName, p),
		label:   label,
		data:    valueFromStructField(fieldName, p),
		viewBuf: &bytes.Buffer{},
	}
}

// domElementSelfClose is a special DOM element which is parsed as a
// self-closing tag and thus needs to be created differently
func domElementSelfClose(e *element) []byte {
	e.viewBuf.Write([]byte(`<div class="input-field col s12">`))
	if e.label != "" {
		e.viewBuf.Write([]byte(`<label class="active" for="` + strings.Join(strings.Split(e.label, " "), "-") + `">` + e.label + `</label>`))
	}
	e.viewBuf.Write([]byte(`<` + e.tagName + ` value="`))
	e.viewBuf.Write([]byte(html.EscapeString(e.data) + `" `))

	for attr, value := range e.attrs {
		e.viewBuf.Write([]byte(attr + `="` + value + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.name + `"`))
	e.viewBuf.Write([]byte(` />`))

	e.viewBuf.Write([]byte(`</div>`))
	return e.viewBuf.Bytes()
}

// domElementCheckbox is a special DOM element which is parsed as a
// checkbox input tag and thus needs to be created differently
func domElementCheckbox(e *element) []byte {
	e.viewBuf.Write([]byte(`<p class="col s6">`))
	e.viewBuf.Write([]byte(`<` + e.tagName + ` `))

	for attr, value := range e.attrs {
		e.viewBuf.Write([]byte(attr + `="` + value + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.name + `"`))
	e.viewBuf.Write([]byte(` /> `))
	if e.label != "" {
		e.viewBuf.Write([]byte(`<label for="` + strings.Join(strings.Split(e.label, " "), "-") + `">` + e.label + `</label>`))
	}
	e.viewBuf.Write([]byte(`</p>`))
	return e.viewBuf.Bytes()
}

// domElement creates a DOM element
func domElement(e *element) []byte {
	e.viewBuf.Write([]byte(`<div class="input-field col s12">`))

	if e.label != "" {
		e.viewBuf.Write([]byte(`<label class="active" for="` + strings.Join(strings.Split(e.label, " "), "-") + `">` + e.label + `</label>`))
	}
	e.viewBuf.Write([]byte(`<` + e.tagName + ` `))

	for attr, value := range e.attrs {
		e.viewBuf.Write([]byte(attr + `="` + string(value) + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.name + `"`))
	e.viewBuf.Write([]byte(` >`))

	e.viewBuf.Write([]byte(html.EscapeString(e.data)))
	e.viewBuf.Write([]byte(`</` + e.tagName + `>`))

	e.viewBuf.Write([]byte(`</div>`))
	return e.viewBuf.Bytes()
}

func domElementWithChildrenSelect(e *element, children []*element) []byte {
	e.viewBuf.Write([]byte(`<div class="input-field col s6">`))

	e.viewBuf.Write([]byte(`<` + e.tagName + ` `))

	for attr, value := range e.attrs {
		e.viewBuf.Write([]byte(attr + `="` + string(value) + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.name + `"`))
	e.viewBuf.Write([]byte(` >`))

	// loop over children and create domElement for each child
	for _, child := range children {
		e.viewBuf.Write(domElement(child))
	}

	e.viewBuf.Write([]byte(`</` + e.tagName + `>`))

	if e.label != "" {
		e.viewBuf.Write([]byte(`<label class="active">` + e.label + `</label>`))
	}

	e.viewBuf.Write([]byte(`</div>`))
	return e.viewBuf.Bytes()
}

func domElementWithChildrenCheckbox(e *element, children []*element) []byte {
	e.viewBuf.Write([]byte(`<` + e.tagName + ` `))

	for attr, value := range e.attrs {
		e.viewBuf.Write([]byte(attr + `="` + value + `" `))
	}

	e.viewBuf.Write([]byte(` >`))

	if e.label != "" {
		e.viewBuf.Write([]byte(`<label class="active">` + e.label + `</label>`))
	}

	// loop over children and create domElement for each child
	for _, child := range children {
		e.viewBuf.Write(domElementCheckbox(child))
	}

	e.viewBuf.Write([]byte(`</` + e.tagName + `><div class="clear padding">&nbsp;</div>`))

	return e.viewBuf.Bytes()
}
