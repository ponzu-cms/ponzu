package config

import "encoding/json"

// GetConfig gets the value of a key in the configuration from the db
func (repo *repository) GetConfig(key string) ([]byte, error) {
	kv := make(map[string]interface{})

	cfg, err := repo.GetAll()
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
