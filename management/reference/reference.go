package reference

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bosssauce/ponzu/management/editor"
	"github.com/bosssauce/ponzu/system/db"
)

// Select returns the []byte of a <select> HTML element plus internal <options> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func Select(fieldName string, p interface{}, attrs map[string]string, contentType, display string) []byte {
	// decode all content type from db into options map
	// map["?type=<contentType>&id=<id>"]t.String()
	options := make(map[string]string)
	jj := db.ContentAll(contentType + "__sorted")

	data := make(map[string]interface{})
	for i := range jj {
		err := json.Unmarshal(jj[i], data)
		if err != nil {
			log.Println("Error decoding into reference handle:", contentType, err)
		}

		k := fmt.Sprintf("?type=%s&id=%d", contentType, data["id"].(int))
		v := data[display].(string)
		options[k] = v
	}

	return editor.Select(fieldName, p, attrs, options)
}
