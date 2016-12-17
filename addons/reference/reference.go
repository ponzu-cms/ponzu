package reference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/bosssauce/ponzu/management/editor"
)

// Referenceable enures there is a way to reference the implenting type from
// within another type's editor and from type-scoped API calls
type Referenceable interface {
	Referenced() []byte
}

// Select returns the []byte of a <select> HTML element plus internal <options> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Select(fieldName string, p interface{}, attrs map[string]string, contentType string, tmpl *template.Template) []byte {
	// decode all content type from db into options map
	// map["?type=<contentType>&id=<id>"]t.String()
	options := make(map[string]string)

	var data []map[string]interface{}
	j := ContentAll(contentType)

	err := json.Unmarshal(j, data)
	if err != nil {
		return nil
	}

	for i := range data {
		k := fmt.Sprintf("?type=%s&id=%s", contentType, data[i]["id"].(string))
		v := &bytes.Buffer{}
		err := tmpl.Execute(v, data[i])
		if err != nil {
			return nil
		}

		options[k] = v.String()
	}

	return editor.Select(fieldName, p, attrs, options)
}

// ContentAll retrives all items from the HTTP API within the provided namespace
func ContentAll(namespace string) []byte {
	endpoint := "http://0.0.0.0:8080/api/contents?type="
	buf := []byte{}
	r := bytes.NewReader(buf)
	req, err := http.NewRequest(http.MethodGet, endpoint+namespace, r)
	if err != nil {
		log.Println("Error creating request for reference from:", namespace)
		return nil
	}

	c := http.Client{}
	res, err := c.Do(req)
	defer res.Body.Close()

	fmt.Println(res, string(buf))

	return buf
}
