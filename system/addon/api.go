// Package addon provides an API for Ponzu addons to interface with the system
package addon

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/bosssauce/ponzu/system/db"
)

// ContentAll retrives all items from the HTTP API within the provided namespace
func ContentAll(namespace string) []byte {
	host := db.ConfigCache("domain")
	port := db.ConfigCache("http_port")
	endpoint := "http://%s:%s/api/contents?type=%s"
	buf := []byte{}
	r := bytes.NewReader(buf)
	url := fmt.Sprintf(endpoint, host, port, namespace)

	req, err := http.NewRequest(http.MethodGet, url, r)
	if err != nil {
		log.Println("Error creating request for reference of:", namespace)
		return nil
	}

	c := http.Client{
		Timeout: time.Duration(time.Second * 5),
	}
	res, err := c.Do(req)
	defer res.Body.Close()

	j, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading request body for reference of:", namespace)
		return nil
	}

	return j
}
