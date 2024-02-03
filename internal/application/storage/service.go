// Package storage provides a re-usable file storage and storage utility for Ponzu
// systems to handle multipart form data.
package storage

import (
	"github.com/fanky5g/ponzu/internal/application"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"github.com/fanky5g/ponzu/internal/domain/services/content"
	"mime/multipart"
)

var ServiceToken application.ServiceToken = "StorageService"
var UploadsEntityName = "uploads"

type service struct {
	client interfaces.StorageClientInterface
	content.Service
}

type Service interface {
	content.Service
	GetAllUploads() ([]item.FileUpload, error)
	GetFileUpload(target string) (*item.FileUpload, error)
	DeleteFile(target string) error
	StoreFiles(files map[string]*multipart.FileHeader) (map[string]string, error)
	interfaces.StaticFileSystemInterface
}

func New(
	contentRepository interfaces.ContentRepositoryInterface,
	configRepository interfaces.ConfigRepositoryInterface,
	searchClient interfaces.SearchClientInterface,
	client interfaces.StorageClientInterface) (Service, error) {
	contentDomainService, err := content.New(contentRepository, configRepository, searchClient)
	if err != nil {
		return nil, err
	}

	if err = searchClient.CreateIndex(UploadsEntityName, &item.FileUpload{}); err != nil {
		return nil, err
	}

	s := &service{
		client:  client,
		Service: contentDomainService,
	}

	return s, nil
}
