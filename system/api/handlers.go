package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/nilslice/cms/content"
	"github.com/nilslice/cms/system/db"
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

func postsHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	// TODO: implement pagination
	// num := q.Get("num")
	// page := q.Get("page")

	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	posts := db.ContentAll(t)
	var all = []json.RawMessage{}
	for _, post := range posts {
		all = append(all, post)
	}

	j, err := fmtJSON(all...)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, j, http.StatusOK)
}

func postHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	id := q.Get("id")
	t := q.Get("type")

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
		data, _ = json.Marshal(map[string]interface{}{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	sendData(res, data, 200)
}

// ResponseFunc ...
type ResponseFunc func(http.ResponseWriter, *http.Request)

// CORS wraps a HandleFunc to response to OPTIONS requests properly
func CORS(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodOptions {
			SendPreflight(res)
			return
		}

		next.ServeHTTP(res, req)
	})
}
