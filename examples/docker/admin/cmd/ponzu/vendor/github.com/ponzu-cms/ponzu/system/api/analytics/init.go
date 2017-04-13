// Package analytics provides the methods to run an analytics reporting system
// for API requests which may be useful to users for measuring access and
// possibly identifying bad actors abusing requests.
package analytics

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

type apiRequest struct {
	URL        string `json:"url"`
	Method     string `json:"http_method"`
	Origin     string `json:"origin"`
	Proto      string `json:"http_protocol"`
	RemoteAddr string `json:"ip_address"`
	Timestamp  int64  `json:"timestamp"`
	External   bool   `json:"external_content"`
}

type apiMetric struct {
	Date   string `json:"date"`
	Total  int    `json:"total"`
	Unique int    `json:"unique"`
}

var (
	store       *bolt.DB
	requestChan chan apiRequest
)

// RANGE determines the number of days ponzu request analytics and metrics are
// stored and displayed within the system
const RANGE = 14

// Record queues an apiRequest for metrics
func Record(req *http.Request) {
	external := strings.Contains(req.URL.Path, "/external/")

	ts := int64(time.Nanosecond) * time.Now().UnixNano() / int64(time.Millisecond)

	r := apiRequest{
		URL:        req.URL.String(),
		Method:     req.Method,
		Origin:     req.Header.Get("Origin"),
		Proto:      req.Proto,
		RemoteAddr: req.RemoteAddr,
		Timestamp:  ts,
		External:   external,
	}

	// put r on buffered requestChan to take advantage of batch insertion in DB
	requestChan <- r
}

// Close exports the abillity to close our db file. Should be called with defer
// after call to Init() from the same place.
func Close() {
	err := store.Close()
	if err != nil {
		log.Println(err)
	}
}

// Init creates a db connection, initializes the db with schema and data and
// sets up the queue/batching channel
func Init() {
	var err error
	store, err = bolt.Open("analytics.db", 0666, nil)
	if err != nil {
		log.Fatalln(err)
	}

	err = store.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("__requests"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("__metrics"))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalln("Error idempotently creating requests bucket in analytics.db:", err)
	}

	requestChan = make(chan apiRequest, 1024*64*runtime.NumCPU())

	go serve()

	if err != nil {
		log.Fatalln(err)
	}
}

func serve() {
	// make timer to notify select to batch request insert from requestChan
	// interval: 30 seconds
	apiRequestTimer := time.NewTicker(time.Second * 30)

	// make timer to notify select to remove analytics older than RANGE days
	// interval: RANGE/2 days
	// TODO: enable analytics backup service to cloud
	pruneThreshold := time.Hour * 24 * RANGE
	pruneDBTimer := time.NewTicker(pruneThreshold / 2)

	for {
		select {
		case <-apiRequestTimer.C:
			err := batchInsert(requestChan)
			if err != nil {
				log.Println(err)
			}

		case <-pruneDBTimer.C:
			err := batchPrune(pruneThreshold)
			if err != nil {
				log.Println(err)
			}

		case <-time.After(time.Second * 30):
			continue
		}
	}
}

// ChartData returns the map containing decoded javascript needed to chart RANGE
// days of data by day
func ChartData() (map[string]interface{}, error) {
	// set thresholds for today and the RANGE-1 days preceding
	times := [RANGE]time.Time{}
	dates := [RANGE]string{}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	ips := [RANGE]map[string]struct{}{}
	for i := range ips {
		ips[i] = make(map[string]struct{})
	}

	total := [RANGE]int{}
	unique := [RANGE]int{}

	for i := range times {
		// subtract 24 * i hours to make days prior
		dur := time.Duration(24 * i * -1)
		day := today.Add(time.Hour * dur)

		// day threshold is [...n-1-i, n-1, n]
		times[len(times)-1-i] = day
		dates[len(times)-1-i] = day.Format("01/02")
	}

	// get api request analytics and metrics from db
	var requests = []apiRequest{}
	currentMetrics := make(map[string]apiMetric)

	err := store.Update(func(tx *bolt.Tx) error {
		m := tx.Bucket([]byte("__metrics"))
		b := tx.Bucket([]byte("__requests"))

		err := m.ForEach(func(k, v []byte) error {
			var metric apiMetric
			err := json.Unmarshal(v, &metric)
			if err != nil {
				log.Println("Error decoding api metric json from analytics db:", err)
				return nil
			}

			// add metric to currentMetrics map
			currentMetrics[metric.Date] = metric

			return nil
		})
		if err != nil {
			return err
		}

		err = b.ForEach(func(k, v []byte) error {
			var r apiRequest
			err := json.Unmarshal(v, &r)
			if err != nil {
				log.Println("Error decoding api request json from analytics db:", err)
				return nil
			}

			// append request to requests for analysis if its timestamp is today
			// or if its day is not already in cache, otherwise delete it
			d := time.Unix(r.Timestamp/1000, 0)
			_, inCache := currentMetrics[d.Format("01/02")]
			if !d.Before(today) || !inCache {
				requests = append(requests, r)
			} else {
				err := b.Delete(k)
				if err != nil {
					return err
				}
			}

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

	// add data to currentMetrics from total and unique
	for i := range dates {
		_, ok := currentMetrics[dates[i]]
		if !ok {
			m := apiMetric{
				Date:   dates[i],
				Total:  total[i],
				Unique: unique[i],
			}

			currentMetrics[dates[i]] = m
		}
	}

	// loop through total and unique to see which dates are accounted for and
	// insert data from metrics array where dates are not
	err = store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__metrics"))

		for i := range dates {
			// populate total and unique with cached data if needed
			if total[i] == 0 {
				total[i] = currentMetrics[dates[i]].Total
			}

			if unique[i] == 0 {
				unique[i] = currentMetrics[dates[i]].Unique
			}

			// check if we need to insert old data into cache - as long as it
			// is not today's data
			if dates[i] != today.Format("01/02") {
				k := []byte(dates[i])
				if b.Get(k) == nil {
					// keep zero counts out of cache in case data is added from
					// other sources
					if currentMetrics[dates[i]].Total != 0 {
						v, err := json.Marshal(currentMetrics[dates[i]])
						if err != nil {
							return err
						}

						err = b.Put(k, v)
						if err != nil {
							return err
						}
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// marshal array counts to js arrays for output to chart
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
		"from":   dates[0],
		"to":     dates[len(dates)-1],
	}, nil
}
