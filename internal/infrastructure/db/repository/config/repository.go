package config

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"github.com/fanky5g/ponzu/internal/util"
	"log"
)

var bucketName = "__config"

type repository struct {
	store *bolt.DB
	cache interfaces.Cache
}

func (repo *repository) Cache() interfaces.Cache {
	return repo.cache
}

// InvalidateCache sets a new Etag for http responses
func (repo *repository) InvalidateCache() error {
	err := repo.PutConfig("etag", util.NewEtag())
	if err != nil {
		return err
	}

	return nil
}

func New(store *bolt.DB) (interfaces.ConfigRepositoryInterface, error) {
	if err := store.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to create storage bucket: %v", bucketName)
	}

	cache, err := NewConfigCache()
	if err != nil {
		return nil, err
	}

	repo := &repository{
		store: store,
		cache: cache,
	}

	// invalidate cache on system start
	err = repo.InvalidateCache()
	if err != nil {
		log.Fatalln("Failed to invalidate cache.", err)
	}

	c, err := repo.GetAll()
	if err != nil {
		return nil, err
	}

	if c == nil {
		c, err = emptyConfig()
		if err != nil {
			return nil, err
		}

		if err = cache.Warm(c); err != nil {
			return nil, err
		}
	}

	return repo, nil
}

func emptyConfig() ([]byte, error) {
	cfg := &entities.Config{}

	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	return data, nil
}
