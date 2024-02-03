package config

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// PutConfig updates a single k/v in the config
func (repo *repository) PutConfig(key string, value interface{}) error {
	kv := make(map[string]interface{})

	c, err := repo.GetAll()
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

	err = repo.SetConfig(data)
	if err != nil {
		return err
	}

	return nil
}
