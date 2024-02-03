package interfaces

import (
	"reflect"
)

type SearchIndexInterface interface {
	Update(id string, data interface{}) error
	Delete(id string) error
	Search(query string, count, offset int) ([]interface{}, error)
}

type Searchable interface {
	GetSearchableAttributes() map[string]reflect.Type
	IndexContent() bool
}
