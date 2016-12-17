package api

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

// Run adds Handlers to default http listener for API
func Run() {
	http.HandleFunc("/api/types", CORS(Record(typesHandler)))

	http.HandleFunc("/api/contents", CORS(Record(contentsHandler)))

	http.HandleFunc("/api/content", CORS(Record(contentHandler)))

	http.HandleFunc("/api/content/external", CORS(Record(externalContentHandler)))
}

// ContentAll retrives all items from the HTTP API within the provided namespace
func ContentAll(namespace string) []byte {
	endpoint := "http://0.0.0.0:8080/api/contents?type="
	buf := []byte{}
	r := bytes.NewReader(buf)
	req, err := http.NewRequest(http.MethodGet, endpoint+namespace, r)
	if err != nil {
		log.Println("Error creating request for reference from:", namespace)
		return nil
	}

	c := http.Client{}
	res, err := c.Do(req)
	defer res.Body.Close()

	fmt.Println(res, string(buf))

	return buf
}
