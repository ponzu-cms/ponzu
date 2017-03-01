package api

import (
	"fmt"
	"log"

	"github.com/ponzu-cms/ponzu/system/item"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func omit(it interface{}, data []byte) ([]byte, error) {
	// is it Omittable
	om, ok := it.(item.Omittable)
	if !ok {
		return data, nil
	}

	return omitFields(om, data, "data")
}

func omitFields(om item.Omittable, data []byte, pathPrefix string) ([]byte, error) {
	// get fields to omit from json data
	fields := om.Omit()

	// remove each field from json, all responses contain json object(s) in top-level "data" array
	n := int(gjson.GetBytes(data, pathPrefix+".#").Int())
	for i := 0; i < n; i++ {
		for k := range fields {
			var err error
			data, err = sjson.DeleteBytes(data, fmt.Sprintf("%s.%d.%s", pathPrefix, i, fields[k]))
			if err != nil {
				log.Println("Erorr omitting field:", fields[k], "from item.Omittable:", om)
				return nil, err
			}
		}
	}

	return data, nil
}
