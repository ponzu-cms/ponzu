package analytics

import (
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"log"
	"runtime"
	"time"
)

// RANGE determines the number of days ponzu request analytics and metrics are
// stored and displayed within the system
const RANGE = 14

var (
	requestChan chan entities.AnalyticsHTTPRequestMetadata
)

// Record queues an apiRequest for metrics
func (s *service) Record(req entities.AnalyticsHTTPRequestMetadata) {
	// put r on buffered requestChan to take advantage of batch insertion in DB
	requestChan <- req
}

func (s *service) StartRecorder(analyticsRepository interfaces.AnalyticsRepositoryInterface) {
	requestChan = make(chan entities.AnalyticsHTTPRequestMetadata, 1024*64*runtime.NumCPU())
	// make timer to notify select to batch request insert from requestChan
	// interval: 30 seconds
	apiRequestTimer := time.NewTicker(time.Second * 30)

	// make timer to notify select to remove analytics older than RANGE days
	// interval: RANGE/2 days
	// TODO: enable analytics backup services to cloud
	pruneThreshold := time.Hour * 24 * RANGE
	pruneDBTimer := time.NewTicker(pruneThreshold / 2)

	for {
		select {
		case <-apiRequestTimer.C:
			var reqs []entities.AnalyticsHTTPRequestMetadata
			batchSize := len(requestChan)

			for i := 0; i < batchSize; i++ {
				reqs = append(reqs, <-requestChan)
			}

			err := analyticsRepository.Insert(reqs)
			if err != nil {
				log.Println(err)
			}

		case <-pruneDBTimer.C:
			err := analyticsRepository.Prune(pruneThreshold)
			if err != nil {
				log.Println(err)
			}

		case <-time.After(time.Second * 30):
			continue
		}
	}
}
