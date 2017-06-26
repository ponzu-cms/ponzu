package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/ponzu-cms/ponzu/system/admin/config"

	"github.com/boltdb/bolt"
	"github.com/gorilla/schema"
)

const (
	// DefaultMaxAge provides a 2592000 second (30-day) cache max-age setting
	DefaultMaxAge = int64(60 * 60 * 24 * 30)
)

var mu = &sync.Mutex{}
var configCache map[string]interface{}

func init() {
	configCache = make(map[string]interface{})
}

// SetConfig sets key:value pairs in the db for configuration settings
func SetConfig(data url.Values) error {
	var j []byte
	err := store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__config"))

		// check for any multi-value fields (ex. checkbox fields)
		// and correctly format for db storage. Essentially, we need
		// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
		fieldOrderValue := make(map[string]map[string][]string)
		for k, v := range data {
			if strings.Contains(k, ".") {
				fo := strings.Split(k, ".")

				// put the order and the field value into map
				field := string(fo[0])
				order := string(fo[1])
				if len(fieldOrderValue[field]) == 0 {
					fieldOrderValue[field] = make(map[string][]string)
				}

				// orderValue is 0:[?type=Thing&id=1]
				orderValue := fieldOrderValue[field]
				orderValue[order] = v
				fieldOrderValue[field] = orderValue

				// discard the post form value with name.N
				data.Del(k)
			}

		}

		// add/set the key & value to the post form in order
		for f, ov := range fieldOrderValue {
			for i := 0; i < len(ov); i++ {
				position := fmt.Sprintf("%d", i)
				fieldValue := ov[position]

				if data.Get(f) == "" {
					for i, fv := range fieldValue {
						if i == 0 {
							data.Set(f, fv)
						} else {
							data.Add(f, fv)
						}
					}
				} else {
					for _, fv := range fieldValue {
						data.Add(f, fv)
					}
				}
			}
		}

		cfg := &config.Config{}
		dec := schema.NewDecoder()
		dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
		dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
		err := dec.Decode(cfg, data)
		if err != nil {
			return err
		}

		// check for "invalidate" value to reset the Etag
		if len(cfg.CacheInvalidate) > 0 && cfg.CacheInvalidate[0] == "invalidate" {
			cfg.Etag = NewEtag()
			cfg.CacheInvalidate = []string{}
		}

		j, err = json.Marshal(cfg)
		if err != nil {
			return err
		}

		err = b.Put([]byte("settings"), j)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// convert json => map[string]interface{}
	var kv map[string]interface{}
	err = json.Unmarshal(j, &kv)
	if err != nil {
		return err
	}

	mu.Lock()
	configCache = kv
	mu.Unlock()

	return nil
}

// Config gets the value of a key in the configuration from the db
func Config(key string) ([]byte, error) {
	kv := make(map[string]interface{})

	cfg, err := ConfigAll()
	if err != nil {
		return nil, err
	}

	if len(cfg) < 1 {
		return nil, nil
	}

	err = json.Unmarshal(cfg, &kv)
	if err != nil {
		return nil, err
	}

	return []byte(kv[key].(string)), nil
}

// ConfigAll gets the configuration from the db
func ConfigAll() ([]byte, error) {
	val := &bytes.Buffer{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__config"))
		if b == nil {
			return fmt.Errorf("Error finding bucket: %s", "__config")
		}
		_, err := val.Write(b.Get([]byte("settings")))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return val.Bytes(), nil
}

// PutConfig updates a single k/v in the config
func PutConfig(key string, value interface{}) error {
	kv := make(map[string]interface{})

	c, err := ConfigAll()
	if err != nil {
		return err
	}

	if c == nil {
		c, err = emptyConfig()
		if err != nil {
			return err
		}
	}

	err = json.Unmarshal(c, &kv)
	if err != nil {
		return err
	}

	// set k/v from params to decoded map
	kv[key] = value

	data := make(url.Values)
	for k, v := range kv {
		switch v.(type) {
		case string:
			data.Set(k, v.(string))

		case []string:
			vv := v.([]string)
			for i := range vv {
				data.Add(k, vv[i])
			}

		default:
			data.Set(k, fmt.Sprintf("%v", v))
		}
	}

	err = SetConfig(data)
	if err != nil {
		return err
	}

	return nil
}

// ConfigCache is a in-memory cache of the Configs for quicker lookups
// 'key' is the JSON tag associated with the config field
func ConfigCache(key string) interface{} {
	mu.Lock()
	val := configCache[key]
	mu.Unlock()

	return val
}

// LoadCacheConfig loads the config into a cache to be accessed by ConfigCache()
func LoadCacheConfig() error {
	c, err := ConfigAll()
	if err != nil {
		return err
	}

	if c == nil {
		c, err = emptyConfig()
		if err != nil {
			return err
		}
	}

	// convert json => map[string]interface{}
	var kv map[string]interface{}
	err = json.Unmarshal(c, &kv)
	if err != nil {
		return err
	}

	mu.Lock()
	configCache = kv
	mu.Unlock()

	return nil
}

func emptyConfig() ([]byte, error) {
	cfg := &config.Config{}

	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	return data, nil
}
