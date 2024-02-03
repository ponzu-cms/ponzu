package interfaces

import (
	"net/http"
)

type StaticFileSystemInterface interface {
	Open(name string) (http.File, error)
}
