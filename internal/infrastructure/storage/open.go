package storage

import (
	"net/http"
)

func (c *client) Open(name string) (http.File, error) {
	return c.s.Open(name)
}
