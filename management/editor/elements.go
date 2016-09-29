package editor

import (
	"bytes"
	"fmt"
	"reflect"
)

type element struct {
	TagName string
	Attrs   map[string]string
	Name    string
	label   string
	data    string
	viewBuf *bytes.Buffer
}

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

// Select returns the []byte of a <select> HTML element plus internal <options> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Select(fieldName string, p interface{}, attrs, options map[string]string) []byte {
	// options are the value attr and the display value, i.e.
	// <option value="{map key}">{map value}</option>

	// find the field value in p to determine if an option is pre-selected
	fieldVal := valueFromStructField(fieldName, p).String()

	// may need to alloc a buffer, as we will probably loop through options
	// and append the []byte from domElement() called for each option
	sel := newElement("select", attrs["label"], fieldName, p, attrs)
	var opts []*element

	// provide a call to action for the select element
	cta := &element{
		TagName: "option",
		Attrs:   map[string]string{"disabled": "true", "selected": "true"},
		data:    "Select an option...",
		viewBuf: &bytes.Buffer{},
	}

	// provide a selection reset (will store empty string in db)
	reset := &element{
		TagName: "option",
		Attrs:   map[string]string{"value": ""},
		data:    "None",
		viewBuf: &bytes.Buffer{},
	}

	opts = append(opts, cta, reset)

	var val string
	for k, v := range options {
		if k == fieldVal {
			val = "true"
		} else {
			val = "false"
		}
		opt := &element{
			TagName: "option",
			Attrs:   map[string]string{"value": k, "selected": val},
			data:    v,
			viewBuf: &bytes.Buffer{},
		}

		// if val is false (unselected option), delete the attr for clarity
		if val == "false" {
			delete(opt.Attrs, "selected")
		}

		opts = append(opts, opt)
	}

	return domElementWithChildren(sel, opts)
}

// Checkbox returns the []byte of a set of <input type="checkbox"> HTML elements
// wrapped in a <div> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Checkbox(fieldName string, p interface{}, attrs, options map[string]string) []byte {
	// options are the value attr and the display value, i.e.
	/*
		<label>
			{map value}
			<input type="checkbox" name="{fieldName}" value="{map key}"/>
		</label>
	*/

	div := newElement("div", attrs["label"], "", p, attrs)
	var opts []*element

	// get the pre-checked options if this is already an existing post
	checkedVals := valueFromStructField(fieldName, p)                         // returns refelct.Value
	checked := checkedVals.Slice(0, checkedVals.Len()).Interface().([]string) // casts reflect.Value to []string

	i := 0
	for k, v := range options {
		// check if k is in the pre-checked values and set to checked
		var val string
		for _, x := range checked {
			if k == x {
				val = "true"
			}
		}

		// create a *element manually using the maodified tagNameFromStructFieldMulti
		// func since this is for a multi-value name
		input := &element{
			TagName: "input",
			Attrs: map[string]string{
				"type":    "checkbox",
				"checked": val,
				"value":   k,
			},
			Name:    tagNameFromStructFieldMulti(fieldName, i, p),
			label:   v,
			data:    "",
			viewBuf: &bytes.Buffer{},
		}

		// if checked == false, delete it from input.Attrs for clarity
		if input.Attrs["checked"] == "" {
			delete(input.Attrs, "checked")
		}

		opts = append(opts, input)
		i++
	}

	return domElementWithChildrenCheckbox(div, opts)
}

// domElementSelfClose is a special DOM element which is parsed as a
// self-closing tag and thus needs to be created differently
func domElementSelfClose(e *element) []byte {
	if e.label != "" {
		e.viewBuf.Write([]byte(`<label>` + e.label))
	}
	e.viewBuf.Write([]byte(`<` + e.TagName + ` value="`))
	e.viewBuf.Write([]byte(e.data + `" `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + value + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.Name + `"`))
	e.viewBuf.Write([]byte(` />`))

	if e.label != "" {
		e.viewBuf.Write([]byte(`</label>`))
	}

	return e.viewBuf.Bytes()
}

// domElementCheckbox is a special DOM element which is parsed as a
// checkbox input tag and thus needs to be created differently
func domElementCheckbox(e *element) []byte {
	if e.label != "" {
		e.viewBuf.Write([]byte(`<label>`))
	}
	e.viewBuf.Write([]byte(`<` + e.TagName + ` `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + value + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.Name + `"`))
	e.viewBuf.Write([]byte(` /> `))

	if e.label != "" {
		e.viewBuf.Write([]byte(e.label + `</label>`))
	}

	return e.viewBuf.Bytes()
}

// domElement creates a DOM element
func domElement(e *element) []byte {
	if e.label != "" {
		e.viewBuf.Write([]byte(`<label>` + e.label))
	}
	e.viewBuf.Write([]byte(`<` + e.TagName + ` `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + string(value) + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.Name + `"`))
	e.viewBuf.Write([]byte(` >`))

	e.viewBuf.Write([]byte(e.data))
	e.viewBuf.Write([]byte(`</` + e.TagName + `>`))

	if e.label != "" {
		e.viewBuf.Write([]byte(`</label>`))
	}

	return e.viewBuf.Bytes()
}

func domElementWithChildren(e *element, children []*element) []byte {
	if e.label != "" {
		e.viewBuf.Write([]byte(`<label>` + e.label))
	}
	e.viewBuf.Write([]byte(`<` + e.TagName + ` `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + string(value) + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.Name + `"`))
	e.viewBuf.Write([]byte(` >`))

	// loop over children and create domElement for each child
	for _, child := range children {
		e.viewBuf.Write(domElement(child))
	}

	e.viewBuf.Write([]byte(`</` + e.TagName + `>`))

	if e.label != "" {
		e.viewBuf.Write([]byte(`</label>`))
	}

	return e.viewBuf.Bytes()
}

func domElementWithChildrenCheckbox(e *element, children []*element) []byte {
	if e.label != "" {
		e.viewBuf.Write([]byte(`<label>` + e.label))
	}
	e.viewBuf.Write([]byte(`<` + e.TagName + ` `))

	for attr, value := range e.Attrs {
		e.viewBuf.Write([]byte(attr + `="` + value + `" `))
	}
	e.viewBuf.Write([]byte(` name="` + e.Name + `"`))
	e.viewBuf.Write([]byte(` >`))

	// loop over children and create domElement for each child
	for _, child := range children {
		e.viewBuf.Write(domElementCheckbox(child))
	}

	e.viewBuf.Write([]byte(`</` + e.TagName + `>`))

	if e.label != "" {
		e.viewBuf.Write([]byte(`</label>`))
	}

	return e.viewBuf.Bytes()
}

func tagNameFromStructField(name string, post interface{}) string {
	// sometimes elements in these environments will not have a name,
	// and thus no tag name in the struct which correlates to it.
	if name == "" {
		return name
	}

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

// due to the format in which gorilla/schema expects form names to be when
// one is associated with multiple values, we need to output the name as such.
// Ex. 'category.0', 'category.1', 'category.2' and so on.
func tagNameFromStructFieldMulti(name string, i int, post interface{}) string {
	tag := tagNameFromStructField(name, post)

	return fmt.Sprintf("%s.%d", tag, i)
}

func valueFromStructField(name string, post interface{}) reflect.Value {
	field := reflect.Indirect(reflect.ValueOf(post)).FieldByName(name)

	return field
}

func newElement(tagName, label, fieldName string, p interface{}, attrs map[string]string) *element {
	return &element{
		TagName: tagName,
		Attrs:   attrs,
		Name:    tagNameFromStructField(fieldName, p),
		label:   label,
		data:    valueFromStructField(fieldName, p).String(),
		viewBuf: &bytes.Buffer{},
	}
}
