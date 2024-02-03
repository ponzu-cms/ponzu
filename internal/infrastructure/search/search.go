package search

import (
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	conf "github.com/fanky5g/ponzu/internal/config"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"log"
	"os"
	"path"
	"strings"
)

var TypeField = "type"
var IndexSuffix = ".index"

type client struct {
	searchPath        string
	indexes           map[string]interfaces.SearchIndexInterface
	contentRepository interfaces.ContentRepositoryInterface
}

// UpdateIndex TODO: only call when an item structure updates (via manual command)
func (c *client) UpdateIndex(entityName string, entityType interface{}) error {
	if err := c.createIndex(entityName, entityType, true); err != nil {
		return err
	}

	searchIndex, err := c.GetIndex(entityName)
	if err != nil {
		return err
	}

	if searchIndex == nil {
		return errors.New("failed to update index")
	}

	go func() {
		entities, err := c.contentRepository.FindAll(entityName)
		if err != nil {
			log.Fatalf("Failed to re-index namespace: %s. Error: %v", entityName, err)
		}

		for _, entity := range entities {
			if err = searchIndex.Update(entity.(item.Identifiable).ItemID(), entity); err != nil {
				log.Fatalf("Failed to index entity: %v", entity)
			}
		}
	}()

	return nil
}

func (c *client) Indexes() (map[string]interfaces.SearchIndexInterface, error) {
	return c.indexes, nil
}

func (c *client) GetIndex(entityName string) (interfaces.SearchIndexInterface, error) {
	if index, ok := c.indexes[entityName]; ok {
		return index, nil
	}

	return nil, fmt.Errorf("index for %s not implemented", entityName)
}

func (c *client) persistIndex(indexPath string, mapping *mapping.IndexMappingImpl) (bleve.Index, error) {
	mapping.TypeField = TypeField
	_, err := os.Stat(indexPath)
	if err == nil {
		if err = os.RemoveAll(indexPath); err != nil {
			return nil, fmt.Errorf("failed to remove existing index: %v", err)
		}
	}

	return bleve.New(indexPath, mapping)
}

func New(contentRepository interfaces.ContentRepositoryInterface) (interfaces.SearchClientInterface, error) {
	searchPath := conf.SearchDir()
	if err := os.MkdirAll(searchPath, os.ModeDir|os.ModePerm); err != nil {
		return nil, err
	}

	c := &client{
		indexes:           make(map[string]interfaces.SearchIndexInterface),
		contentRepository: contentRepository,
		searchPath:        searchPath,
	}

	managedTypes := contentRepository.Types()
	searchDirItems, err := os.ReadDir(searchPath)
	if err != nil {
		return nil, err
	}

	if len(searchDirItems) > 0 {
		for _, searchDirItem := range searchDirItems {
			if searchDirItem.IsDir() {
				entityName := strings.TrimSuffix(searchDirItem.Name(), IndexSuffix)
				if _, ok := managedTypes[entityName]; ok {
					searchIndex, err := c.getExistingIndex(path.Join(searchPath, searchDirItem.Name()), false)
					if err != nil {
						return nil, err
					}

					if searchIndex != nil {
						log.Printf("Search index %s initialized\n", entityName)
						c.indexes[entityName] = searchIndex
					}
				}

			}
		}
	}

	return c, nil
}
