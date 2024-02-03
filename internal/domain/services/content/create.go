package content

import (
	"errors"
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"github.com/gofrs/uuid"
	"strings"
)

func (s *service) CreateContent(entityType string, entity interface{}) (string, error) {
	identifiable, ok := entity.(item.Identifiable)
	if !ok {
		return "", errors.New("item does not implement identifiable interface")
	}

	if identifiable.UniqueID().IsNil() {
		// add UUID to data for use in embedded Item
		uid, err := uuid.NewV4()
		if err != nil {
			return "", err
		}

		entity.(item.Identifiable).SetUniqueID(uid)
	}

	if sluggable, ok := entity.(item.Sluggable); ok && sluggable.ItemSlug() == "" {
		slug, err := item.Slug(entity.(item.Identifiable))
		if err != nil {
			return "", err
		}

		slug, err = s.repository.UniqueSlug(slug)
		if err != nil {
			return "", err
		}

		entity.(item.Sluggable).SetSlug(slug)
	}

	id, err := s.repository.SetEntity(entityType, entity)
	if err != nil {
		return "", err
	}

	if err = s.configRepository.InvalidateCache(); err != nil {
		return "", err
	}

	var index interfaces.SearchIndexInterface
	index, err = s.searchClient.GetIndex(s.getEntityType(entityType))
	if err != nil {
		return "", fmt.Errorf("failed to index %s for search", entityType)
	}

	if err = index.Update(id, entity); err != nil {
		return "", err
	}

	return fmt.Sprint(id), nil
}

func (s *service) getEntityType(target string) string {
	return strings.Split(target, ":")[0]
}
