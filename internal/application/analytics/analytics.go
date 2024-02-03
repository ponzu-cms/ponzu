// Package analytics provides the methods to run an analytics reporting system
// for API requests which may be useful to users for measuring access and
// possibly identifying bad actors abusing requests.
package analytics

import (
	"github.com/fanky5g/ponzu/internal/application"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
)

var ServiceToken application.ServiceToken = "AnalyticsService"

type service struct {
	repository interfaces.AnalyticsRepositoryInterface
}

type Service interface {
	StartRecorder(analyticsRepository interfaces.AnalyticsRepositoryInterface)
	Record(req entities.AnalyticsHTTPRequestMetadata)
	GetChartData() (map[string]interface{}, error)
}

func New(repository interfaces.AnalyticsRepositoryInterface) (Service, error) {
	return &service{repository: repository}, nil
}
