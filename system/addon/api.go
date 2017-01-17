// Package addon provides an API for Ponzu addons to interface with the system
package addon

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ponzu-cms/ponzu/system/db"
)

// QueryOptions is a mirror of the same struct in db package and are re-declared
// here only to make the API simpler for the caller
type QueryOptions db.QueryOptions

// ContentAll retrives all items from the HTTP API within the provided namespace
func ContentAll(namespace string) []byte {
	host := db.ConfigCache("domain").(string)
	port := db.ConfigCache("http_port").(string)
	endpoint := "http://%s:%s/api/contents?type=%s&count=-1"
	URL := fmt.Sprintf(endpoint, host, port, namespace)

	j, err := Get(URL)
	if err != nil {
		log.Println("Error in ContentAll for reference HTTP request:", URL)
		return nil
	}

	return j
}

// Query retrieves a set of content from the HTTP API  based on options
// and returns the total number of content in the namespace and the content
func Query(namespace string, opts QueryOptions) []byte {
	host := db.ConfigCache("domain").(string)
	port := db.ConfigCache("http_port").(string)
	endpoint := "http://%s:%s/api/contents?type=%s&count=%d&offset=%d&order=%s"
	URL := fmt.Sprintf(endpoint, host, port, namespace, opts.Count, opts.Offset, opts.Order)

	j, err := Get(URL)
	if err != nil {
		log.Println("Error in Query for reference HTTP request:", URL)
		return nil
	}

	return j
}

// Get is a helper function to make a HTTP call from an addon
func Get(endpoint string) ([]byte, error) {
	buf := []byte{}
	r := bytes.NewReader(buf)

	req, err := http.NewRequest(http.MethodGet, endpoint, r)
	if err != nil {
		log.Println("Error creating reference HTTP request:", endpoint)
		return nil, err
	}

	c := http.Client{
		Timeout: time.Duration(time.Second * 5),
	}
	res, err := c.Do(req)
	if err != nil {
		log.Println("Error making reference HTTP request:", endpoint)
		return nil, err
	}
	defer res.Body.Close()

	j, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading body for reference HTTP request:", endpoint)
		return nil, err
	}

	return j, nil
}
