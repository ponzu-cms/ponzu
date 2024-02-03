package search

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/application"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
)

var ContentSearchService application.ServiceToken = "ContentSearchService"
var UploadSearchService application.ServiceToken = "UploadSearchService"

type service struct {
	client interfaces.SearchClientInterface
}

type Service interface {
	Search(entityName, query string, count, offset int) ([]interface{}, error)
}

func New(client interfaces.SearchClientInterface) (Service, error) {
	return &service{client: client}, nil
}

// Search conducts a search and returns a set of Ponzu "targets", Type:ID pairs,
// and an error. If there is no search index for the typeName (Type) provided,
// db.ErrNoIndex will be returned as the error
func (s *service) Search(entityName, query string, count, offset int) ([]interface{}, error) {
	index, err := s.client.GetIndex(entityName)
	if err != nil {
		return nil, fmt.Errorf("failed to get index for entity: %s", entityName)
	}

	return index.Search(query, count, offset)
}
