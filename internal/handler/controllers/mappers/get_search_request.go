package mappers

import (
	"errors"
	"github.com/fanky5g/ponzu/internal/handler/controllers/resources"
	"net/url"
	"strconv"
)

func GetSearchRequest(qs url.Values) (*resources.SearchRequestDto, error) {
	q, err := url.QueryUnescape(qs.Get("q"))
	if err != nil {
		return nil, err
	}

	// q must be set
	if q == "" {
		return nil, errors.New("query is required")
	}

	count, err := strconv.Atoi(qs.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if qs.Get("count") == "" {
			count = 10
		} else {
			return nil, err
		}
	}

	offset, err := strconv.Atoi(qs.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if qs.Get("offset") == "" {
			offset = 0
		} else {
			return nil, err
		}
	}

	return &resources.SearchRequestDto{
		Query:  q,
		Count:  count,
		Offset: offset,
	}, nil
}
