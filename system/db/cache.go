package db

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CacheControl sets the default cache policy on static asset responses
func CacheControl(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		etag := ConfigCache("etag")
		policy := fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 60*60*24*30)
		res.Header().Add("Etag", etag)
		res.Header().Add("Cache-Control", policy)

		if match := res.Header().Get("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				fmt.Println("matched etag")
				res.WriteHeader(http.StatusNotModified)
				return
			}

			fmt.Println("checked, no match")
		}

		next.ServeHTTP(res, req)
	})
}

// NewEtag generates a new Etag for response caching
func NewEtag() string {
	now := fmt.Sprintf("%d", time.Now().Unix())
	etag := base64.StdEncoding.EncodeToString([]byte(now))

	return etag
}

// InvalidateCache sets a new Etag for http responses
func InvalidateCache() error {
	kv := make(map[string]interface{})

	c, err := ConfigAll()
	if err != nil {
		return err
	}

	err = json.Unmarshal(c, &kv)
	if err != nil {
		return err
	}

	kv["etag"] = NewEtag()

	data := make(url.Values)
	for k, v := range kv {
		switch v.(type) {
		case string:
			data.Set(k, v.(string))
		case []string:
			vv := v.([]string)
			for i := range vv {
				if i == 0 {
					data.Set(k, vv[i])
				} else {
					data.Add(k, vv[i])
				}
			}
		}
	}

	err = SetConfig(data)
	if err != nil {

	}

	return nil
}
