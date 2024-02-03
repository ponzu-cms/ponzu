package storage

import (
	"github.com/fanky5g/ponzu/internal/config"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"net/http"
)

type client struct {
	s interfaces.StaticFileSystemInterface
}

func New() (interfaces.StorageClientInterface, error) {
	s, err := NewLocalStaticFileSystem(http.Dir(config.UploadDir()))
	if err != nil {
		return nil, err
	}

	return &client{s: s}, nil
}
