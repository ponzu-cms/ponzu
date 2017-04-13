package editor

import (
	"bytes"
	"html"
	"log"
	"strings"
)

// Element is a basic struct for representing DOM elements
type Element struct {
	TagName string
	Attrs   map[string]string
	Name    string
	Label   string
	Data    string
	ViewBuf *bytes.Buffer
}

// NewElement returns an Element with Name and Data already processed from the
// fieldName and content interface provided
func NewElement(tagName, label, fieldName string, p interface{}, attrs map[string]string) *Element {
	return &Element{
		TagName: tagName,
		Attrs:   attrs,
		Name:    TagNameFromStructField(fieldName, p),
		Label:   label,
		Data:    ValueFromStructField(fieldName, p),
		ViewBuf: &bytes.Buffer{},
	}
}

// DOMElementSelfClose is a special DOM element which is parsed as a
// self-closing tag and thus needs to be created differently
func DOMElementSelfClose(e *Element) []byte {
	_, err := e.ViewBuf.WriteString(`<div class="input-field col s12">`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementSelfClose")
		return nil
	}

	if e.Label != "" {
		_, err = e.ViewBuf.WriteString(
			`<label class="active" for="` +
				strings.Join(strings.Split(e.Label, " "), "-") + `">` + e.Label +
				`</label>`)
		if err != nil {
			log.Println("Error writing HTML string to buffer: DOMElementSelfClose")
			return nil
		}
	}

	_, err = e.ViewBuf.WriteString(`<` + e.TagName + ` value="`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementSelfClose")
		return nil
	}

	_, err = e.ViewBuf.WriteString(html.EscapeString(e.Data) + `" `)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementSelfClose")
		return nil
	}

	for attr, value := range e.Attrs {
		_, err := e.ViewBuf.WriteString(attr + `="` + value + `" `)
		if err != nil {
			log.Println("Error writing HTML string to buffer: DOMElementSelfClose")
			return nil
		}
	}
	_, err = e.ViewBuf.WriteString(` name="` + e.Name + `" />`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementSelfClose")
		return nil
	}

	_, err = e.ViewBuf.WriteString(`</div>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementSelfClose")
		return nil
	}

	return e.ViewBuf.Bytes()
}

// DOMElementCheckbox is a special DOM element which is parsed as a
// checkbox input tag and thus needs to be created differently
func DOMElementCheckbox(e *Element) []byte {
	_, err := e.ViewBuf.WriteString(`<p class="col s6">`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementCheckbox")
		return nil
	}

	_, err = e.ViewBuf.WriteString(`<` + e.TagName + ` `)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementCheckbox")
		return nil
	}

	for attr, value := range e.Attrs {
		_, err := e.ViewBuf.WriteString(attr + `="` + value + `" `)
		if err != nil {
			log.Println("Error writing HTML string to buffer: DOMElementCheckbox")
			return nil
		}
	}
	_, err = e.ViewBuf.WriteString(` name="` + e.Name + `" />`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementCheckbox")
		return nil
	}

	if e.Label != "" {
		_, err = e.ViewBuf.WriteString(
			`<label for="` +
				strings.Join(strings.Split(e.Label, " "), "-") + `">` +
				e.Label + `</label>`)
		if err != nil {
			log.Println("Error writing HTML string to buffer: DOMElementCheckbox")
			return nil
		}
	}

	_, err = e.ViewBuf.WriteString(`</p>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementCheckbox")
		return nil
	}

	return e.ViewBuf.Bytes()
}

// DOMElement creates a DOM element
func DOMElement(e *Element) []byte {
	_, err := e.ViewBuf.WriteString(`<div class="input-field col s12">`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElement")
		return nil
	}

	if e.Label != "" {
		_, err = e.ViewBuf.WriteString(
			`<label class="active" for="` +
				strings.Join(strings.Split(e.Label, " "), "-") + `">` + e.Label +
				`</label>`)
		if err != nil {
			log.Println("Error writing HTML string to buffer: DOMElement")
			return nil
		}
	}

	_, err = e.ViewBuf.WriteString(`<` + e.TagName + ` `)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElement")
		return nil
	}

	for attr, value := range e.Attrs {
		_, err = e.ViewBuf.WriteString(attr + `="` + string(value) + `" `)
		if err != nil {
			log.Println("Error writing HTML string to buffer: DOMElement")
			return nil
		}
	}
	_, err = e.ViewBuf.WriteString(` name="` + e.Name + `" >`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElement")
		return nil
	}

	_, err = e.ViewBuf.WriteString(html.EscapeString(e.Data))
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElement")
		return nil
	}

	_, err = e.ViewBuf.WriteString(`</` + e.TagName + `>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElement")
		return nil
	}

	_, err = e.ViewBuf.WriteString(`</div>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElement")
		return nil
	}

	return e.ViewBuf.Bytes()
}

func DOMElementWithChildrenSelect(e *Element, children []*Element) []byte {
	_, err := e.ViewBuf.WriteString(`<div class="input-field col s6">`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementWithChildrenSelect")
		return nil
	}

	_, err = e.ViewBuf.WriteString(`<` + e.TagName + ` `)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementWithChildrenSelect")
		return nil
	}

	for attr, value := range e.Attrs {
		_, err = e.ViewBuf.WriteString(attr + `="` + value + `" `)
		if err != nil {
			log.Println("Error writing HTML string to buffer: DOMElementWithChildrenSelect")
			return nil
		}
	}
	_, err = e.ViewBuf.WriteString(` name="` + e.Name + `" >`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementWithChildrenSelect")
		return nil
	}

	// loop over children and create DOMElement for each child
	for _, child := range children {
		_, err = e.ViewBuf.Write(DOMElement(child))
		if err != nil {
			log.Println("Error writing HTML DOMElement to buffer: DOMElementWithChildrenSelect")
			return nil
		}
	}

	_, err = e.ViewBuf.WriteString(`</` + e.TagName + `>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementWithChildrenSelect")
		return nil
	}

	if e.Label != "" {
		_, err = e.ViewBuf.WriteString(`<label class="active">` + e.Label + `</label>`)
		if err != nil {
			log.Println("Error writing HTML string to buffer: DOMElementWithChildrenSelect")
			return nil
		}
	}

	_, err = e.ViewBuf.WriteString(`</div>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementWithChildrenSelect")
		return nil
	}

	return e.ViewBuf.Bytes()
}

func DOMElementWithChildrenCheckbox(e *Element, children []*Element) []byte {
	_, err := e.ViewBuf.WriteString(`<` + e.TagName + ` `)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementWithChildrenCheckbox")
		return nil
	}

	for attr, value := range e.Attrs {
		_, err = e.ViewBuf.WriteString(attr + `="` + value + `" `)
		if err != nil {
			log.Println("Error writing HTML string to buffer: DOMElementWithChildrenCheckbox")
			return nil
		}
	}

	_, err = e.ViewBuf.WriteString(` >`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementWithChildrenCheckbox")
		return nil
	}

	if e.Label != "" {
		_, err = e.ViewBuf.WriteString(`<label class="active">` + e.Label + `</label>`)
		if err != nil {
			log.Println("Error writing HTML string to buffer: DOMElementWithChildrenCheckbox")
			return nil
		}
	}

	// loop over children and create DOMElement for each child
	for _, child := range children {
		_, err = e.ViewBuf.Write(DOMElementCheckbox(child))
		if err != nil {
			log.Println("Error writing HTML DOMElementCheckbox to buffer: DOMElementWithChildrenCheckbox")
			return nil
		}
	}

	_, err = e.ViewBuf.WriteString(`</` + e.TagName + `><div class="clear padding">&nbsp;</div>`)
	if err != nil {
		log.Println("Error writing HTML string to buffer: DOMElementWithChildrenCheckbox")
		return nil
	}

	return e.ViewBuf.Bytes()
}
