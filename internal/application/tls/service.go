package tls

import (
	"github.com/fanky5g/ponzu/internal/application"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
)

var ServiceToken application.ServiceToken = "TlsService"

type service struct {
	configRepository interfaces.ConfigRepositoryInterface
}

type Service interface {
	Enable()
	EnableDev()
}

func New(repository interfaces.ConfigRepositoryInterface) (Service, error) {
	return &service{configRepository: repository}, nil
}
