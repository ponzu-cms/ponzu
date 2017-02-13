package api

import (
	"log"

	"github.com/ponzu-cms/ponzu/system/item"

	"github.com/tidwall/sjson"
)

func omit(it interface{}, data []byte) ([]byte, error) {
	// is it Omittable
	om, ok := it.(item.Omittable)
	if !ok {
		return data, nil
	}

	return omitFields(om, data, "data.0.")
}

func omitFields(om item.Omittable, data []byte, pathPrefix string) ([]byte, error) {
	// get fields to omit from json data
	fields := om.Omit()

	// remove each field from json, all responses contain json object(s) in top-level "data" array
	var omitted = data
	for i := range fields {
		var err error
		omitted, err = sjson.DeleteBytes(omitted, pathPrefix+fields[i])
		if err != nil {
			log.Println("Erorr omitting field:", fields[i], "from item.Omittable:", om)
			return nil, err
		}
	}

	return omitted, nil
}
