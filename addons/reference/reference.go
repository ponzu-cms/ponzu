package reference

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bosssauce/ponzu/content"
	"github.com/bosssauce/ponzu/management/editor"
	"github.com/bosssauce/ponzu/system/db"
)

// New returns the []byte of a <select> HTML element plus internal <options> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func New(fieldName string, p interface{}, attrs map[string]string, contentType, fmtString string) []byte {
	ct, ok := content.Types[contentType]
	if !ok {
		log.Println("Cannot reference an invalid content type:", contentType)
		return nil
	}

	// get a handle to the underlying interface type for decoding
	t := ct()

	fmt.Println(t)

	// // decode all content type from db into options map
	// // map["?type=<contentType>&id=<id>"]t.String()
	options := make(map[string]string)
	jj := db.ContentAll(contentType + "__sorted")

	for i := range jj {
		err := json.Unmarshal(jj[i], t)
		if err != nil {
			log.Println("Error decoding into reference handle:", contentType, err)
		}

		// make sure it is an Identifiable
		item, ok := t.(content.Identifiable)
		if !ok {
			log.Println("Cannot use type", contentType, "as reference as it does not implement content.Identifiable")
			return nil
		}

		k := fmt.Sprintf("?type=%s&id=%d", contentType, item.ItemID())
		v := item.String()
		options[k] = v
	}

	options[""] = contentType + "Content loading..."

	return editor.Select(fieldName, p, attrs, options)
}

/*
<script>
        // fmtString = "{name} - ( Age: {age} | Power: {power} )"
		// $(function() {
		// 	var API = '/api/contents?type=` + contentType + `';
		// 	var select = $('select[name="` + name + `"]);

		// 	$.getJSON(API, function(resp, status) {
		// 		if (status !== '200' || status !== '304') {
		// 			console.log('Error loading Reference for', '` + contentType + `')
		// 			return
		// 		}

		// 		var data = resp.data,
		// 			options = [],
		// 			re = /{(.*?)}/g,
		// 			tmpl = '` + fmtString + `'
		// 			tags = tmpl.match(re),
		// 			keys = [];

		// 		// get keys from tags ({x} -> x)
		// 		for (var i = 0; i < tags.length; i++) {
		// 			var key = tags[i].slice(1, tags[i].length-1);
		// 			keys.push(key);
		// 		}

		// 		// create options as objects of "?type=<contentType>&id=<id>":displayName
		// 		for (var i = 0; i < data.length; i++) {

		// 		}
		// 	});
		// });
	</script>
*/
