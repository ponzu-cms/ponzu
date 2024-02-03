package storage

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
)

func (s *service) DeleteFile(target string) error {
	fileUpload, err := s.GetContent(UploadsEntityName, target)
	if err != nil {
		return err
	}

	f, ok := fileUpload.(*item.FileUpload)
	if !ok {
		return fmt.Errorf("failed to delete file: invalid item matched: %T", fileUpload)
	}

	if err = s.client.Delete(f.Path); err != nil {
		return fmt.Errorf("failed to delete from file store: %v", err)
	}

	return s.Service.DeleteContent(UploadsEntityName, target)
}
