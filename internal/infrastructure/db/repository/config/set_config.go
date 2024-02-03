package config

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"github.com/fanky5g/ponzu/internal/util"
	"github.com/gorilla/schema"
	"net/url"
	"strings"
)

// SetConfig sets key:value pairs in the db for configuration settings
// TODO: set config is also doing exactly the same thing as we do in mappers get entity
func (repo *repository) SetConfig(data url.Values) error {
	var j []byte
	err := repo.store.Update(func(tx *bolt.Tx) error {
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

		cfg := &entities.Config{}
		dec := schema.NewDecoder()
		dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
		dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
		err := dec.Decode(cfg, data)
		if err != nil {
			return err
		}

		// check for "invalidate" value to reset the Etag
		if len(cfg.CacheInvalidate) > 0 && cfg.CacheInvalidate[0] == "invalidate" {
			cfg.Etag = util.NewEtag()
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

	return repo.Cache().Warm(j)
}
