package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/nilslice/cms/content"
	"github.com/nilslice/cms/system/db"
)

// Run adds Handlers to default http listener for API
func Run() {
	http.HandleFunc("/api/types", func(res http.ResponseWriter, req *http.Request) {
		var types = []string{}
		for t := range content.Types {
			types = append(types, string(t))
		}

		j, err := toJSON(types)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write(j)
	})

	http.HandleFunc("/api/posts", func(res http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()
		t := q.Get("type")
		// TODO: implement pagination
		// num := q.Get("num")
		// page := q.Get("page")

		if t == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		posts := db.GetAll(t)
		var all = []json.RawMessage{}
		for _, post := range posts {
			all = append(all, post)
		}

		j, err := fmtJSON(all...)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write(j)
	})

	http.HandleFunc("/api/post", func(res http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()
		id := q.Get("id")
		t := q.Get("type")

		if t == "" || id == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		post, err := db.Get(t + ":" + id)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		j, err := fmtJSON(json.RawMessage(post))
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write(j)
	})

}

func fmtJSON(data ...json.RawMessage) ([]byte, error) {
	var msg = []json.RawMessage{}
	for _, d := range data {
		msg = append(msg, d)
	}

	resp := map[string][]json.RawMessage{
		"data": msg,
	}

	var buf = &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := enc.Encode(resp)
	if err != nil {
		log.Println("Failed to encode data to JSON:", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func toJSON(data []string) ([]byte, error) {
	var buf = &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	resp := map[string][]string{
		"data": data,
	}

	err := enc.Encode(resp)
	if err != nil {
		log.Println("Failed to encode data to JSON:", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func wrapJSON(json []byte) []byte {
	var buf = &bytes.Buffer{}
	buf.Write([]byte("{data:"))
	buf.Write(json)
	buf.Write([]byte("}"))

	return buf.Bytes()
}
