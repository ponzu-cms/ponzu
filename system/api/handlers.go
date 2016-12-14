package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bosssauce/ponzu/content"
	"github.com/bosssauce/ponzu/system/api/analytics"
	"github.com/bosssauce/ponzu/system/db"
)

func typesHandler(res http.ResponseWriter, req *http.Request) {
	var types = []string{}
	for t := range content.Types {
		types = append(types, string(t))
	}

	j, err := toJSON(types)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, j, http.StatusOK)
}

func contentsHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, ok := content.Types[t]; !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	count, err := strconv.Atoi(q.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if q.Get("count") == "" {
			count = 10
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	offset, err := strconv.Atoi(q.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if q.Get("offset") == "" {
			offset = 0
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	order := strings.ToLower(q.Get("order")) // string: sort order of posts by timestamp ASC / DESC (DESC default)
	if order != "asc" {
		order = "desc"
	}

	opts := db.QueryOptions{
		Count:  count,
		Offset: offset,
		Order:  order,
	}

	_, bb := db.Query(t+"__sorted", opts)
	var result = []json.RawMessage{}
	for i := range bb {
		result = append(result, bb[i])
	}

	j, err := fmtJSON(result...)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, j, http.StatusOK)
}

func contentHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	id := q.Get("id")
	t := q.Get("type")
	slug := q.Get("slug")

	if slug != "" {
		contentHandlerBySlug(res, req)
		return
	}

	if _, ok := content.Types[t]; !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if t == "" || id == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	post, err := db.Content(t + ":" + id)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err := fmtJSON(json.RawMessage(post))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, j, http.StatusOK)
}

func contentHandlerBySlug(res http.ResponseWriter, req *http.Request) {
	slug := req.URL.Query().Get("slug")

	// lookup type:id by slug key in __contentIndex
	post, err := db.ContentBySlug(slug)
	if err != nil {
		log.Println("Error finding content by slug:", slug, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err := fmtJSON(json.RawMessage(post))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, j, http.StatusOK)
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
	buf.Write([]byte(`{"data":`))
	buf.Write(json)
	buf.Write([]byte(`}`))

	return buf.Bytes()
}

// sendData() should be used any time you want to communicate
// data back to a foreign client
func sendData(res http.ResponseWriter, data []byte, code int) {
	res.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(code)
	res.Write(data)
}

// SendPreflight is used to respond to a cross-origin "OPTIONS" request
func SendPreflight(res http.ResponseWriter) {
	res.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.WriteHeader(200)
	return
}

// SendJSON returns a Response to a client as JSON
func SendJSON(res http.ResponseWriter, j map[string]interface{}) {
	var data []byte
	var err error

	data, err = json.Marshal(j)
	if err != nil {
		log.Println(err)
		data, _ = json.Marshal(map[string]interface{}{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	sendData(res, data, 200)
}

// CORS wraps a HandleFunc to response to OPTIONS requests properly
func CORS(next http.HandlerFunc) http.HandlerFunc {
	return db.CacheControl(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodOptions {
			SendPreflight(res)
			return
		}

		next.ServeHTTP(res, req)
	}))
}

// Record wraps a HandleFunc to record API requests for analytical purposes
func Record(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		go analytics.Record(req)

		next.ServeHTTP(res, req)
	})
}
