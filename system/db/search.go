package db

import (
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

// Search tracks all search indices to use throughout system
var Search map[string]bleve.Index

// Searchable ...
type Searchable interface {
	SearchMapping() *mapping.IndexMappingImpl
}

func init() {
	Search = make(map[string]bleve.Index)
}

// MapIndex creates the mapping for a type and tracks the index to be used within
// the system for adding/deleting/checking data
func MapIndex(typeName string) error {
	// TODO: type assert for Searchable, get configuration (which can be overridden)
	// by Ponzu user if defines own SearchMapping()

	mapping := bleve.NewIndexMapping()
	mapping.StoreDynamic = false
	idxName := typeName + ".index"
	var idx bleve.Index

	// check if index exists, use it or create new one
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	searchPath := filepath.Join(pwd, "search")

	err = os.MkdirAll(searchPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	idxPath := filepath.Join(searchPath, idxName)
	if _, err = os.Stat(idxPath); os.IsNotExist(err) {
		idx, err = bleve.New(idxPath, mapping)
		if err != nil {
			return err
		}
	} else {
		idx, err = bleve.Open(idxPath)
		if err != nil {
			return err
		}
	}

	// add the type name to the index and track the index
	Search[typeName] = idx

	return nil
}
