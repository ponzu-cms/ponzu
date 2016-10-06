package db

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/boltdb/bolt"
	"github.com/gorilla/schema"
	"github.com/nilslice/cms/system/admin/config"
)

// SetConfig sets key:value pairs in the db for configuration settings
func SetConfig(data url.Values) error {
	err := store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("_config"))

		cfg := &config.Config{}
		dec := schema.NewDecoder()
		dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
		dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
		err := dec.Decode(cfg, data)
		if err != nil {
			return err
		}

		j, err := json.Marshal(cfg)
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

	return nil
}

// Config gets the value of a key in the configuration from the db
func Config(key string) ([]byte, error) {
	kv := make(map[string]interface{})

	cfg, err := ConfigAll()
	if err != nil {
		return nil, err
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
		b := tx.Bucket([]byte("_config"))
		val.Write(b.Get([]byte("settings")))

		return nil
	})
	if err != nil {
		return nil, err
	}

	return val.Bytes(), nil
}
