// Package analytics provides the methods to run an analytics reporting system
// for API requests which may be useful to users for measuring access and
// possibly identifying bad actors abusing requests.
package analytics

import (
	"log"
	"net/http"
	"strings"
	"time"

	"runtime"

	"github.com/boltdb/bolt"
)

type apiRequest struct {
	URL        string `json:"url"`
	Method     string `json:"http_method"`
	Origin     string `json:"origin"`
	Proto      string `json:"http_protocol"`
	RemoteAddr string `json:"ip_address"`
	Timestamp  int64  `json:"timestamp"`
	External   bool   `json:"external"`
}

var (
	store      *bolt.DB
	recordChan chan apiRequest
)

// Record queues an apiRequest for metrics
func Record(req *http.Request) {
	external := strings.Contains(req.URL.Path, "/external/")

	r := apiRequest{
		URL:        req.URL.String(),
		Method:     req.Method,
		Origin:     req.Header.Get("Origin"),
		Proto:      req.Proto,
		RemoteAddr: req.RemoteAddr,
		Timestamp:  time.Now().Unix() * 1000,
		External:   external,
	}

	// put r on buffered recordChan to take advantage of batch insertion in DB
	recordChan <- r
}

// Close exports the abillity to close our db file. Should be called with defer
// after call to Init() from the same place.
func Close() {
	err := store.Close()
	if err != nil {
		log.Println(err)
	}
}

// Init creates a db connection, should run an initial prune of old data, and
// sets up the queue/batching channel
func Init() {
	var err error
	store, err = bolt.Open("analytics.db", 0666, nil)
	if err != nil {
		log.Fatalln(err)
	}

	recordChan = make(chan apiRequest, 1024*64*runtime.NumCPU())

	go serve()

	if err != nil {
		log.Fatalln(err)
	}
}

func serve() {
	// make timer to notify select to batch request insert from recordChan
	// interval: 30 seconds
	apiRequestTimer := time.NewTicker(time.Second * 30)

	// make timer to notify select to remove old analytics
	// interval: 2 weeks
	// TODO: enable analytics backup service to cloud
	pruneDBTimer := time.NewTicker(time.Hour * 24 * 14)

	for {
		select {
		case <-apiRequestTimer.C:
			var reqs []apiRequest
			batchSize := len(recordChan)

			for i := 0; i < batchSize; i++ {
				reqs = append(reqs, <-recordChan)
			}

			err := batchInsert(reqs)
			if err != nil {
				log.Println(err)
			}

		case <-pruneDBTimer.C:

		default:
		}
	}
}

// Week returns the []byte of javascript needed to chart a week of data by day
func Week() {
	// get request analytics for today and the 6 days preceeding

}
