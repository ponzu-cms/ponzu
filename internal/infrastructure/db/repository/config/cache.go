package config

import (
	"encoding/json"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"sync"
)

type configCache struct {
	mutex *sync.Mutex
	store map[string]interface{}
}

// GetByKey gets a cached config by 'key'
func (cache *configCache) GetByKey(key string) interface{} {
	cache.mutex.Lock()
	val := cache.store[key]
	cache.mutex.Unlock()

	return val
}

// Warm loads the config into a cache to be accessed by key
func (cache *configCache) Warm(value []byte) error {
	// convert json => map[string]interface{}
	var kv map[string]interface{}
	err := json.Unmarshal(value, &kv)
	if err != nil {
		return err
	}

	cache.mutex.Lock()
	cache.store = kv
	cache.mutex.Unlock()

	return nil
}

func NewConfigCache() (interfaces.Cache, error) {
	return &configCache{
		store: make(map[string]interface{}),
		mutex: &sync.Mutex{},
	}, nil
}
