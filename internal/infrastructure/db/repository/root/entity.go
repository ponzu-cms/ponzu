package root

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"io"
	"strings"
)

func (repo *repository) MarshalEntity(target string, value io.Reader) (interface{}, error) {
	t, err := repo.getEntityBuilder(target)
	if err != nil {
		return nil, err
	}

	entity := t()
	if err := json.NewDecoder(value).Decode(&entity); err != nil {
		return nil, err
	}

	return entity, nil
}

func (repo *repository) getEntityBuilder(target string) (item.EntityBuilder, error) {
	if strings.Contains(target, "__") {
		spec := strings.Split(target, "__")
		target = spec[0]
	}

	parts := strings.Split(target, ":")
	entityType := parts[0]

	fn, ok := repo.entityMap[entityType]
	if !ok {
		return nil, fmt.Errorf(item.ErrTypeNotRegistered.Error(), entityType)
	}

	return fn, nil
}

func (repo *repository) Types() map[string]item.EntityBuilder {
	return repo.entityMap
}

func (repo *repository) CreateEntityStore(entityName string, entityType interface{}) error {
	if err := repo.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(entityName))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(entityName + "__sorted"))
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return repo.Sort(entityName)
}
