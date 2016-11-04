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

	"encoding/json"

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

		case <-time.After(time.Second * 30):
			continue
		}
	}
}

// Week returns the map containing decoded javascript needed to chart a week of data by day
func Week() (map[string]interface{}, error) {
	// set thresholds for today and the 6 days preceeding
	times := [7]time.Time{}
	dates := [7]string{}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	for i := range times {
		// subtract 24 * i hours to make days prior
		dur := time.Duration(24 * i * -1)
		day := today.Add(time.Hour * dur)

		// day threshold is [...n-1-i, n-1, n]
		times[len(times)-1-i] = day
		dates[len(times)-1-i] = day.Format("01/02")
	}

	// get api request analytics from db
	var requests = []apiRequest{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("requests"))

		err := b.ForEach(func(k, v []byte) error {
			var r apiRequest
			err := json.Unmarshal(v, &r)
			if err != nil {
				log.Println("Error decoding json from analytics db:", err)
				return nil
			}

			requests = append(requests, r)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	ips := [7]map[string]struct{}{}
	for i := range ips {
		ips[i] = make(map[string]struct{})
	}

	total := [7]int{}
	unique := [7]int{}

CHECK_REQUEST:
	for i := range requests {
		ts := time.Unix(requests[i].Timestamp/1000, 0)

		for j := range times {
			// if on today, there will be no next iteration to set values for
			// day prior so all valid requests belong to today
			if j == len(times)-1 {
				if ts.After(times[j]) || ts.Equal(times[j]) {
					// do all record keeping
					total[j]++

					if _, ok := ips[j][requests[i].RemoteAddr]; !ok {
						unique[j]++
						ips[j][requests[i].RemoteAddr] = struct{}{}
					}

					continue CHECK_REQUEST
				}
			}

			if ts.Equal(times[j]) {
				// increment total count for current time threshold (day)
				total[j]++

				// if no IP found for current threshold, increment unique and record IP
				if _, ok := ips[j][requests[i].RemoteAddr]; !ok {
					unique[j]++
					ips[j][requests[i].RemoteAddr] = struct{}{}
				}

				continue CHECK_REQUEST
			}

			if ts.Before(times[j]) {
				// check if older than earliest threshold
				if j == 0 {
					continue CHECK_REQUEST
				}

				// increment total count for previous time threshold (day)
				total[j-1]++

				// if no IP found for day prior, increment unique and record IP
				if _, ok := ips[j-1][requests[i].RemoteAddr]; !ok {
					unique[j-1]++
					ips[j-1][requests[i].RemoteAddr] = struct{}{}
				}
			}
		}
	}

	jsUnique, err := json.Marshal(unique)
	if err != nil {
		return nil, err
	}

	jsTotal, err := json.Marshal(total)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"dates":  dates,
		"unique": string(jsUnique),
		"total":  string(jsTotal),
	}, nil
}
