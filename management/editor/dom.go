package editor

import (
	"bytes"
	"html"
	"log"
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
	_, err := e.viewBuf.WriteString(`<div class="input-field col s12">`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementSelfClose")
		return nil
	}

	if e.label != "" {
		_, err = e.viewBuf.WriteString(
			`<label class="active" for="` +
				strings.Join(strings.Split(e.label, " "), "-") + `">` + e.label +
				`</label>`)
		if err != nil {
			log.Println("Error writing HTML string to buffer: domElementSelfClose")
			return nil
		}
	}

	_, err = e.viewBuf.WriteString(`<` + e.tagName + ` value="`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementSelfClose")
		return nil
	}

	_, err = e.viewBuf.WriteString(html.EscapeString(e.data) + `" `)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementSelfClose")
		return nil
	}

	for attr, value := range e.attrs {
		_, err := e.viewBuf.WriteString(attr + `="` + value + `" `)
		if err != nil {
			log.Println("Error writing HTML string to buffer: domElementSelfClose")
			return nil
		}
	}
	_, err = e.viewBuf.WriteString(` name="` + e.name + `" />`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementSelfClose")
		return nil
	}

	_, err = e.viewBuf.WriteString(`</div>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementSelfClose")
		return nil
	}

	return e.viewBuf.Bytes()
}

// domElementCheckbox is a special DOM element which is parsed as a
// checkbox input tag and thus needs to be created differently
func domElementCheckbox(e *element) []byte {
	_, err := e.viewBuf.WriteString(`<p class="col s6">`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementCheckbox")
		return nil
	}

	_, err = e.viewBuf.WriteString(`<` + e.tagName + ` `)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementCheckbox")
		return nil
	}

	for attr, value := range e.attrs {
		_, err := e.viewBuf.WriteString(attr + `="` + value + `" `)
		if err != nil {
			log.Println("Error writing HTML string to buffer: domElementCheckbox")
			return nil
		}
	}
	_, err = e.viewBuf.WriteString(` name="` + e.name + `" />`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementCheckbox")
		return nil
	}

	if e.label != "" {
		_, err = e.viewBuf.WriteString(
			`<label for="` +
				strings.Join(strings.Split(e.label, " "), "-") + `">` +
				e.label + `</label>`)
		if err != nil {
			log.Println("Error writing HTML string to buffer: domElementCheckbox")
			return nil
		}
	}

	_, err = e.viewBuf.WriteString(`</p>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementCheckbox")
		return nil
	}

	return e.viewBuf.Bytes()
}

// domElement creates a DOM element
func domElement(e *element) []byte {
	_, err := e.viewBuf.WriteString(`<div class="input-field col s12">`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElement")
		return nil
	}

	if e.label != "" {
		_, err = e.viewBuf.WriteString(
			`<label class="active" for="` +
				strings.Join(strings.Split(e.label, " "), "-") + `">` + e.label +
				`</label>`)
		if err != nil {
			log.Println("Error writing HTML string to buffer: domElement")
			return nil
		}
	}

	_, err = e.viewBuf.WriteString(`<` + e.tagName + ` `)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElement")
		return nil
	}

	for attr, value := range e.attrs {
		_, err = e.viewBuf.WriteString(attr + `="` + string(value) + `" `)
		if err != nil {
			log.Println("Error writing HTML string to buffer: domElement")
			return nil
		}
	}
	_, err = e.viewBuf.WriteString(` name="` + e.name + `" >`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElement")
		return nil
	}

	_, err = e.viewBuf.WriteString(html.EscapeString(e.data))
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElement")
		return nil
	}

	_, err = e.viewBuf.WriteString(`</` + e.tagName + `>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElement")
		return nil
	}

	_, err = e.viewBuf.WriteString(`</div>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElement")
		return nil
	}

	return e.viewBuf.Bytes()
}

func domElementWithChildrenSelect(e *element, children []*element) []byte {
	_, err := e.viewBuf.WriteString(`<div class="input-field col s6">`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementWithChildrenSelect")
		return nil
	}

	_, err = e.viewBuf.WriteString(`<` + e.tagName + ` `)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementWithChildrenSelect")
		return nil
	}

	for attr, value := range e.attrs {
		_, err = e.viewBuf.WriteString(attr + `="` + value + `" `)
		if err != nil {
			log.Println("Error writing HTML string to buffer: domElementWithChildrenSelect")
			return nil
		}
	}
	_, err = e.viewBuf.WriteString(` name="` + e.name + `" >`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementWithChildrenSelect")
		return nil
	}

	// loop over children and create domElement for each child
	for _, child := range children {
		_, err = e.viewBuf.Write(domElement(child))
		if err != nil {
			log.Println("Error writing HTML domElement to buffer: domElementWithChildrenSelect")
			return nil
		}
	}

	_, err = e.viewBuf.WriteString(`</` + e.tagName + `>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementWithChildrenSelect")
		return nil
	}

	if e.label != "" {
		_, err = e.viewBuf.WriteString(`<label class="active">` + e.label + `</label>`)
		if err != nil {
			log.Println("Error writing HTML string to buffer: domElementWithChildrenSelect")
			return nil
		}
	}

	_, err = e.viewBuf.WriteString(`</div>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementWithChildrenSelect")
		return nil
	}

	return e.viewBuf.Bytes()
}

func domElementWithChildrenCheckbox(e *element, children []*element) []byte {
	_, err := e.viewBuf.WriteString(`<` + e.tagName + ` `)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementWithChildrenCheckbox")
		return nil
	}

	for attr, value := range e.attrs {
		_, err = e.viewBuf.WriteString(attr + `="` + value + `" `)
		if err != nil {
			log.Println("Error writing HTML string to buffer: domElementWithChildrenCheckbox")
			return nil
		}
	}

	_, err = e.viewBuf.WriteString(` >`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementWithChildrenCheckbox")
		return nil
	}

	if e.label != "" {
		_, err = e.viewBuf.WriteString(`<label class="active">` + e.label + `</label>`)
		if err != nil {
			log.Println("Error writing HTML string to buffer: domElementWithChildrenCheckbox")
			return nil
		}
	}

	// loop over children and create domElement for each child
	for _, child := range children {
		_, err = e.viewBuf.Write(domElementCheckbox(child))
		if err != nil {
			log.Println("Error writing HTML domElementCheckbox to buffer: domElementWithChildrenCheckbox")
			return nil
		}
	}

	_, err = e.viewBuf.WriteString(`</` + e.tagName + `><div class="clear padding">&nbsp;</div>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: domElementWithChildrenCheckbox")
		return nil
	}

	return e.viewBuf.Bytes()
}
