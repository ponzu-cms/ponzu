package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

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

// sendData should be used any time you want to communicate
// data back to a foreign client
func sendData(res http.ResponseWriter, req *http.Request, data []byte) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Vary", "Accept-Encoding")

	_, err := res.Write(data)
	if err != nil {
		log.Println("Error writing to response in sendData")
	}
}
