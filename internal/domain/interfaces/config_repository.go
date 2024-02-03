package interfaces

import "net/url"

type ConfigRepositoryInterface interface {
	Cacheable
	SetConfig(data url.Values) error
	GetConfig(key string) ([]byte, error)
	GetAll() ([]byte, error)
	PutConfig(key string, value interface{}) error
}
