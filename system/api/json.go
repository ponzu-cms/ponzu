package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
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

// Convert JSON at data to url.Values
func jsonToURLValues(data *[]byte) (*url.Values, error) {
	var formData url.Values
	formData = make(map[string][]string)
	jsonData := make(map[string]interface{})

	if err := json.NewDecoder(bytes.NewReader(*data)).Decode(&jsonData); err != nil {
		return nil, fmt.Errorf("error decoding post form %v", err)
	}

	for k, v := range jsonData {
		switch v.(type) {
		case string:
			formData.Set(k, v.(string))
		case []string:
			for i, v := range v.([]string) {
				if i == 0 {
					formData.Set(k, v)
				} else {
					formData.Add(k, v)
				}
			}
		default:
			return nil, fmt.Errorf("jsonData bad format")
		}
	}
	ts := fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UnixNano()/int64(time.Millisecond))
	formData.Set("timestamp", ts)
	formData.Set("updated", ts)
	return &formData, nil
}
