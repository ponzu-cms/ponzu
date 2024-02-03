package interfaces

import (
	"io"
	"net/http"
)

type StorageClientInterface interface {
	// Save saves a file to storage. Make sure to close file after it's written
	Save(fileName string, file io.ReadCloser) (string, int64, error)
	Delete(path string) error
	Open(name string) (http.File, error)
}
