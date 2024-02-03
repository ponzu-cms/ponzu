package storage

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"mime/multipart"
	"strings"
	"time"
)

// StoreFiles stores file uploads at paths like /YYYY/MM/filename.ext
func (s *service) StoreFiles(files map[string]*multipart.FileHeader) (map[string]string, error) {
	paths := make(map[string]string)
	for name, fileHeader := range files {
		nameParts := strings.Split(name, ":")
		fileName := nameParts[0]
		if len(nameParts) > 1 {
			fileName = nameParts[1]
		}

		f, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("could not open file for uploading: %v", err)
		}

		u, size, err := s.client.Save(fileName, f)
		if err != nil {
			return nil, err
		}

		paths[fileName] = u
		if err = s.storeFileInfo(size, fileName, u, fileHeader); err != nil {
			return nil, err
		}
	}

	return paths, nil
}

func (s *service) storeFileInfo(size int64, filename, urlPath string, file *multipart.FileHeader) error {
	ts := int64(time.Nanosecond) * time.Now().UTC().UnixNano() / int64(time.Millisecond)
	entity := &item.FileUpload{
		Name:          filename,
		Path:          urlPath,
		ContentLength: size,
		ContentType:   file.Header.Get("Content-Type"),
		Item: item.Item{
			Timestamp: ts,
			Updated:   ts,
		},
	}

	if _, err := s.Service.CreateContent(UploadsEntityName, entity); err != nil {
		return fmt.Errorf("error saving file storage record to database: %v", err)
	}

	return nil
}
