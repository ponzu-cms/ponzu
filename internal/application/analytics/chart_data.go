package analytics

import (
	"encoding/json"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"time"
)

// GetChartData returns the map containing decoded javascript needed to chart RANGE
// days of data by day
func (s *service) GetChartData() (map[string]interface{}, error) {
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

	currentMetrics, err := s.repository.GetMetrics()
	if err != nil {
		return nil, err
	}

	requests, err := s.repository.GetRequestMetadata(today, currentMetrics)
	if err != nil {
		return nil, err
	}

CheckRequest:
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

					continue CheckRequest
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

				continue CheckRequest
			}

			if ts.Before(times[j]) {
				// check if older than the earliest threshold
				if j == 0 {
					continue CheckRequest
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
			m := entities.AnalyticsMetric{
				Date:   dates[i],
				Total:  total[i],
				Unique: unique[i],
			}

			currentMetrics[dates[i]] = m
		}
	}

	// loop through total and unique to see which dates are accounted for and
	// insert data from metrics array where dates are not
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
			metric, err := s.repository.GetMetric(k)
			if err != nil {
				return nil, err
			}

			if metric == nil {
				// keep zero counts out of cache in case data is added from
				// other sources
				if currentMetrics[dates[i]].Total != 0 {
					v, err := json.Marshal(currentMetrics[dates[i]])
					if err != nil {
						return nil, err
					}

					err = s.repository.SetMetric(k, v)
					if err != nil {
						return nil, err
					}
				}
			}
		}
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
