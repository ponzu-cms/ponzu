package api

import (
	"log"
	"net/http"

	"github.com/ponzu-cms/ponzu/system/item"

	"github.com/tidwall/sjson"
)

func omit(it interface{}, res http.ResponseWriter, req *http.Request, data *[]byte) ([]byte, error) {
	// is it Omittable
	om, ok := it.(item.Omittable)
	if !ok {
		return *data, nil
	}

	// get fields to omit from json data
	fields := om.Omit()

	// remove each field from json, all responses contain json object(s) in top-level "data" array
	var omitted []byte
	for i := range fields {
		var err error
		omitted, err = sjson.DeleteBytes(*data, "data."+fields[i])
		if err != nil {
			log.Println("Erorr omitting field:", fields[i], "from item.Omittable:", it)
			return nil, err
		}
	}

	return omitted, nil
}
