package api

import (
	"log"
	"net/http"

	"github.com/ponzu-cms/ponzu/system/item"

	"github.com/tidwall/gjson"
	"golang.org/x/net/http2"
)

func push(res http.ResponseWriter, req *http.Request, pt func() interface{}, data []byte) {
	// Push(target string, opts *PushOptions) error
	if pusher, ok := res.(http.Pusher); ok {
		if p, ok := pt().(item.Pushable); ok {
			// get fields to pull values from data
			fields := p.Push()

			// parse values from data to push
			values := gjson.GetManyBytes(data, fields...)

			// push all values from Pushable items' fields
			for i := range values {
				val := values[i]
				val.ForEach(func(k, v gjson.Result) bool {
					if v.String() == "null" {
						return true
					}

					// check that the push is not to its parent URL
					if v.String() == (req.URL.Path + "?" + req.URL.RawQuery) {
						return true
					}

					err := pusher.Push(v.String(), nil)
					// check for error, "http2: recursive push not allowed"
					// and return, suppressing a log message
					if err != nil && err.Error() == http2.ErrRecursivePush.Error() {
						return true
					}
					if err != nil {
						log.Println("Error during Push of value:", v.String(), err)
					}

					return true
				})
			}
		}
	}

}
