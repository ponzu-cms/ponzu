package content

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
)

func (s *service) GetContent(entityType, entityId string) (interface{}, error) {
	// TODO: repository layer accept entityType and entityId
	target := fmt.Sprintf("%s:%s", entityType, entityId)
	return s.repository.FindOneByTarget(target)
}

func (s *service) GetAll(namespace string) ([]interface{}, error) {
	return s.repository.FindAll(namespace)
}

func (s *service) Query(namespace string, options interfaces.QueryOptions) (int, []interface{}, error) {
	return s.repository.Query(namespace, options)
}
