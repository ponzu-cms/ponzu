package storage

import (
	"net/http"
)

func (s *service) Open(name string) (http.File, error) {
	return s.client.Open(name)
}
