package interfaces

import "github.com/fanky5g/ponzu/internal/domain/entities/item"

// QueryOptions holds options for a query
type QueryOptions struct {
	Count  int
	Offset int
	Order  string
}

type CRUDInterface interface {
	// SetEntity performs both Create and Update function
	SetEntity(entityType string, entity interface{}) (string, error)
	DeleteEntity(entityId string) error
	FindByTarget(targets []string) ([]interface{}, error)
	FindOneByTarget(target string) (interface{}, error)
	FindOneBySlug(slug string) (string, interface{}, error)
	FindAll(namespace string) ([]interface{}, error)
	Query(namespace string, opts QueryOptions) (int, []interface{}, error)
}

type EntityIdentifierInterface interface {
	UniqueSlug(slug string) (string, error)
	IsValidID(id string) bool
	NextIDSequence(entityType string) (string, error)
}

type ContentRepositoryInterface interface {
	CreateEntityStore(entityName string, entityType interface{}) error
	CRUDInterface
	EntityIdentifierInterface
	Types() map[string]item.EntityBuilder
}
