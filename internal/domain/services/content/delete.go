package content

import (
	"fmt"
)

func (s *service) DeleteContent(entityType, entityId string) error {
	// TODO: repository layer accept entityType and entityId
	target := fmt.Sprintf("%s:%s", entityType, entityId)
	if err := s.repository.DeleteEntity(target); err != nil {
		return err
	}

	index, err := s.searchClient.GetIndex(s.getEntityType(target))
	if err != nil {
		return fmt.Errorf("failed to delete search index: %v", err)
	}

	return index.Delete(target)
}
