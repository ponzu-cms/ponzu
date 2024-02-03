package config

import (
	"github.com/fanky5g/ponzu/internal/application"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"net/url"
)

var ServiceToken application.ServiceToken = "ConfigService"

type service struct {
	repository interfaces.ConfigRepositoryInterface
}

type Service interface {
	GetAppName() (string, error)
	GetStringValue(key string) (string, error)
	GetCacheStringValue(key string) (string, error)
	GetCacheBoolValue(key string) (bool, error)
	SetConfig(data url.Values) error
	GetAll() ([]byte, error)
}

func (s *service) GetStringValue(key string) (string, error) {
	value, err := s.repository.GetConfig(key)
	if err != nil {
		return "", err
	}

	if value == nil {
		return "", nil
	}

	return string(value), nil
}

func (s *service) GetCacheStringValue(key string) (string, error) {
	value := s.repository.Cache().GetByKey(key)
	if value == nil {
		return "", nil
	}

	if stringValue, ok := value.(string); ok {
		return stringValue, nil
	}

	return "", nil
}

func (s *service) GetCacheBoolValue(key string) (bool, error) {
	value := s.repository.Cache().GetByKey(key)
	if value == nil {
		return false, nil
	}

	if boolValue, ok := value.(bool); ok {
		return boolValue, nil
	}

	return false, nil
}

func (s *service) SetConfig(data url.Values) error {
	return s.repository.SetConfig(data)
}

func (s *service) GetAll() ([]byte, error) {
	return s.repository.GetAll()
}

// GetAppName
// TODO: store app name in cache
func (s *service) GetAppName() (string, error) {
	return s.GetStringValue("name")
}

func New(repository interfaces.ConfigRepositoryInterface) (Service, error) {
	return &service{repository: repository}, nil
}
