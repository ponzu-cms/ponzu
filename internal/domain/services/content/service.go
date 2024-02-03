package content

import (
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
)

type service struct {
	repository       interfaces.ContentRepositoryInterface
	configRepository interfaces.ConfigRepositoryInterface
	searchClient     interfaces.SearchClientInterface
}

type Service interface {
	DeleteContent(entityType, entityId string) error
	CreateContent(entityType string, content interface{}) (string, error)
	GetContent(entityType, entityId string) (interface{}, error)
	Query(entityType string, options interfaces.QueryOptions) (int, []interface{}, error)
	GetAll(entityType string) ([]interface{}, error)
}

func New(
	contentRepository interfaces.ContentRepositoryInterface,
	configRepository interfaces.ConfigRepositoryInterface,
	searchClient interfaces.SearchClientInterface,
) (Service, error) {
	s := &service{
		repository:       contentRepository,
		configRepository: configRepository,
		searchClient:     searchClient,
	}

	return s, nil
}
