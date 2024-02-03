package interfaces

import (
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"time"
)

type AnalyticsRepositoryInterface interface {
	Insert(reqs []entities.AnalyticsHTTPRequestMetadata) error
	Prune(threshold time.Duration) error
	GetMetrics() (map[string]entities.AnalyticsMetric, error)
	GetMetric(key []byte) ([]byte, error)
	SetMetric(key, value []byte) error
	GetRequestMetadata(t time.Time, metrics map[string]entities.AnalyticsMetric) ([]entities.AnalyticsHTTPRequestMetadata, error)
}
