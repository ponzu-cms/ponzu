package search

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"log"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)

func (c *client) getExistingIndex(indexPath string, failOnMissingIndex bool) (interfaces.SearchIndexInterface, error) {
	entityName := strings.TrimSuffix(path.Base(indexPath), IndexSuffix)
	if searchIndex, ok := c.indexes[entityName]; ok {
		return searchIndex, nil
	}

	var index bleve.Index
	index, err := bleve.Open(indexPath)
	if err != nil {
		if failOnMissingIndex {
			log.Printf("Invalid index in search path: %v\n", err)
			return nil, err
		}

		return nil, nil
	}

	index.SetName(entityName)

	var searchIndex interfaces.SearchIndexInterface
	searchIndex, err = NewSearchIndex(entityName, index, c.contentRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to create search index: %v", err)
	}

	return searchIndex, nil
}

func (c *client) createIndex(entityName string, entityType interface{}, overwrite bool) error {
	entity, ok := entityType.(interfaces.Searchable)
	if !ok {
		return fmt.Errorf("entity %s does not implement Searchable interface", entityName)
	}

	if !entity.IndexContent() {
		fmt.Printf("%s is not searchable. Skipping index creation\n", entityName)
		return nil
	}

	idxName := fmt.Sprintf("%s%s", entityName, IndexSuffix)
	idxPath := filepath.Join(c.searchPath, idxName)

	existingIndex, err := c.getExistingIndex(idxPath, overwrite)
	if err != nil {
		if !overwrite {
			return err
		}
	}

	if existingIndex != nil && !overwrite {
		c.indexes[idxName] = existingIndex
		return nil
	}

	typeField := bleve.NewTextFieldMapping()
	typeField.Analyzer = keyword.Name

	entityMapping := bleve.NewDocumentMapping()
	entityMapping.AddFieldMappingsAt(TypeField, typeField)

	for fieldName, fieldType := range entity.GetSearchableAttributes() {
		switch fieldType.Kind() {
		case reflect.String:
			fieldMapping := bleve.NewTextFieldMapping()
			entityMapping.AddFieldMappingsAt(fieldName, fieldMapping)
		default:
			return fmt.Errorf("unsupported field type: %s", fieldType.Kind())
		}
	}

	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultMapping = entityMapping

	index, err := c.persistIndex(idxPath, indexMapping)
	if err != nil {
		return fmt.Errorf("failed to build index: %v", err)
	}

	searchIndex, err := NewSearchIndex(entityName, index, c.contentRepository)
	if err != nil {
		return err
	}

	c.indexes[entityName] = searchIndex
	return nil
}

// CreateIndex TODO: only call when creating an entity (via manual command)
func (c *client) CreateIndex(entityName string, entityType interface{}) error {
	return c.createIndex(entityName, entityType, false)
}
