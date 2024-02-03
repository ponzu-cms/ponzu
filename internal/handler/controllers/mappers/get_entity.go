package mappers

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"github.com/gorilla/schema"
	"net/url"
	"strings"
	"time"
)

func getGenericEntity(t item.EntityBuilder, data url.Values) (interface{}, error) {
	entity := t()
	ts := data.Get("timestamp")
	up := data.Get("updated")

	// create a timestamp if one was not set
	if ts == "" {
		ts = fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UTC().UnixNano()/int64(time.Millisecond))
		data.Set("timestamp", ts)
	}

	if up == "" {
		data.Set("updated", ts)
	}

	// check for any multi-value fields (ex. checkbox fields)
	// and correctly format for db storage. Essentially, we need
	// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
	fieldOrderValue := make(map[string]map[string][]string)
	for k, v := range data {
		if strings.Contains(k, ".") {
			fo := strings.Split(k, ".")

			// put the order and the field value into map
			field := fo[0]
			order := fo[1]
			if len(fieldOrderValue[field]) == 0 {
				fieldOrderValue[field] = make(map[string][]string)
			}

			// orderValue is 0:[?type=Thing&id=1]
			orderValue := fieldOrderValue[field]
			orderValue[order] = v
			fieldOrderValue[field] = orderValue

			// discard the entity form value with name.N
			data.Del(k)
		}
	}

	// add/set the key & value to the entity form in order
	for f, ov := range fieldOrderValue {
		for i := 0; i < len(ov); i++ {
			position := fmt.Sprintf("%d", i)
			fieldValue := ov[position]

			if data.Get(f) == "" {
				for i, fv := range fieldValue {
					if i == 0 {
						data.Set(f, fv)
					} else {
						data.Add(f, fv)
					}
				}
			} else {
				for _, fv := range fieldValue {
					data.Add(f, fv)
				}
			}
		}
	}

	dec := schema.NewDecoder()
	dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
	dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
	err := dec.Decode(entity, data)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func GetEntityFromFormData(entityType string, data url.Values) (interface{}, error) {
	// find the content type and decode values into it
	t, ok := item.Types[entityType]
	if !ok {
		return nil, fmt.Errorf(item.ErrTypeNotRegistered.Error(), entityType)
	}

	return getGenericEntity(t, data)
}
