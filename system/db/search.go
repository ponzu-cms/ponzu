package db

import (
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve"
)

// Search tracks all search indices to use throughout system
var Search map[string]bleve.Index

func init() {
	Search = make(map[string]bleve.Index)
}

// MapIndex creates the mapping for a type and tracks the index to be used within
// the system for adding/deleting/checking data
func MapIndex(typeName string) error {
	mapping := bleve.NewIndexMapping()
	mapping.StoreDynamic = false
	idxFile := typeName + ".index"
	var idx bleve.Index

	// check if index exists, use it or create new one
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err = os.Stat(filepath.Join(pwd, idxFile)); os.IsNotExist(err) {
		idx, err = bleve.New(idxFile, mapping)
		if err != nil {
			return err
		}
	} else {
		idx, err = bleve.Open(idxFile)
		if err != nil {
			return err
		}
	}

	// add the type name to the index and track the index
	Search[typeName] = idx

	return nil
}
