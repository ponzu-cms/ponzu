package editor

import (
	"bytes"
	"reflect"
)

// Input returns the []byte of an <input> HTML element with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Input(fieldName string, p interface{}, attrs map[string]string) []byte {
	e := newElement("input", attrs["label"], fieldName, p, attrs)

	return domElementSelfClose(e)
}

// Textarea returns the []byte of a <textarea> HTML element with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Textarea(fieldName string, p interface{}, attrs map[string]string) []byte {
	e := newElement("textarea", attrs["label"], fieldName, p, attrs)

	return domElement(e)
}

type element struct {
	TagName string
	Attrs   map[string]string
	Name    string
	label   string
	data    string
	viewBuf *bytes.Buffer
}

// domElementSelfClose is a special DOM element which is parsed as a
// self-closing tag and thus needs to be created differently
func domElementSelfClose(e *element) []byte {
	if e.label != "" {
		e.viewBuf.Write([]byte(`<label>` + e.label + `</label>`))
	}
	e.viewBuf.Write([]byte(`<` + e.TagName + ` value="`))
	e.viewBuf.Write([]byte(e.data + `" `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + value + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.Name + `"`))
	e.viewBuf.Write([]byte(` />`))

	return e.viewBuf.Bytes()
}

// domElement creates a DOM element
func domElement(e *element) []byte {
	if e.label != "" {
		e.viewBuf.Write([]byte(`<label>` + e.label + `</label>`))
	}
	e.viewBuf.Write([]byte(`<` + e.TagName + ` `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + string(value) + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.Name + `"`))
	e.viewBuf.Write([]byte(` >`))

	e.viewBuf.Write([]byte(e.data))
	e.viewBuf.Write([]byte(`</` + e.TagName + `>`))

	return e.viewBuf.Bytes()
}

func tagNameFromStructField(name string, post interface{}) string {
	field, ok := reflect.TypeOf(post).Elem().FieldByName(name)
	if !ok {
		panic("Couldn't get struct field for: " + name + ". Make sure you pass the right field name to editor field elements.")
	}

	tag, ok := field.Tag.Lookup("json")
	if !ok {
		panic("Couldn't get json struct tag for: " + name + ". Struct fields for content types must have 'json' tags.")
	}

	return tag
}

func valueFromStructField(name string, post interface{}) string {
	field := reflect.Indirect(reflect.ValueOf(post)).FieldByName(name)

	return field.String()
}

func newElement(tagName, label, fieldName string, p interface{}, attrs map[string]string) *element {
	return &element{
		TagName: tagName,
		Attrs:   attrs,
		Name:    tagNameFromStructField(fieldName, p),
		label:   label,
		data:    valueFromStructField(fieldName, p),
		viewBuf: &bytes.Buffer{},
	}
}
