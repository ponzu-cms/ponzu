package storage

import (
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
)

func (s *service) GetFileUpload(key string) (*item.FileUpload, error) {
	file, err := s.Service.GetContent(UploadsEntityName, key)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, nil
	}

	return file.(*item.FileUpload), nil
}

func (s *service) GetAllUploads() ([]item.FileUpload, error) {
	files, err := s.Service.GetAll(UploadsEntityName)
	if err != nil {
		return nil, err
	}

	f := make([]item.FileUpload, len(files))
	for i, file := range files {
		f[i] = file.(item.FileUpload)
	}

	return f, nil
}
