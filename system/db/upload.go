package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ponzu-cms/ponzu/system/item"

	"strconv"

	"github.com/boltdb/bolt"
	"github.com/gorilla/schema"
	uuid "github.com/satori/go.uuid"
)

// SetUpload stores information about files uploaded to the system
func SetUpload(target string, data url.Values) (int, error) {
	parts := strings.Split(target, ":")
	if parts[0] != "__uploads" {
		return 0, fmt.Errorf("cannot call SetUpload with target type: %s", parts[0])
	}
	pid := parts[1]

	if data.Get("uuid") == "" {
		// set new UUID for upload
		data.Set("uuid", uuid.NewV4().String())
	}

	if data.Get("slug") == "" {
		// create slug based on filename and timestamp/updated fields
		slug := data.Get("name")
		slug, err := checkSlugForDuplicate(slug)
		if err != nil {
			return 0, err
		}
		data.Set("slug", slug)
	}

	ts := fmt.Sprintf("%d", time.Now().Unix()*1000)
	if data.Get("timestamp") == "" {
		data.Set("timestamp", ts)
	}

	data.Set("updated", ts)

	// store in database
	var id int64
	err := store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("__uploads"))
		if err != nil {
			return err
		}

		if pid == "-1" {
			// get sequential ID for item
			id, err := b.NextSequence()
			if err != nil {
				return err
			}
			data.Set("id", fmt.Sprintf("%d", id))
		} else {
			id, err = strconv.ParseInt(pid, 10, 64)
			if err != nil {
				return err
			}
		}

		file := &item.FileUpload{}
		dec := schema.NewDecoder()
		dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
		dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
		err = dec.Decode(file, data)
		if err != nil {
			return err
		}

		// marshal data to json for storage
		j, err := json.Marshal(file)
		if err != nil {
			return err
		}

		err = b.Put([]byte(data.Get("id")), j)
		if err != nil {
			return err
		}

		// add slug to __contentIndex for lookup
		b, err = tx.CreateBucketIfNotExists([]byte("__contentIndex"))
		if err != nil {
			return err
		}

		k := []byte(data.Get("slug"))
		v := []byte(fmt.Sprintf("%s:%d", "__uploads", id))
		err = b.Put(k, v)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Upload returns the value for an upload by its target (__uploads:{id})
func Upload(target string) ([]byte, error) {
	val := &bytes.Buffer{}
	parts := strings.Split(target, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid target for upload: %s", target)
	}

	id := []byte(parts[1])

	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__uploads"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		j := b.Get(id)
		_, err := val.Write(j)
		return err
	})

	return val.Bytes(), err
}

// UploadBySlug returns the value for an upload by its slug
func UploadBySlug(slug string) ([]byte, error) {
	val := &bytes.Buffer{}
	// get target from __contentIndex or return nil if not exists
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__contentIndex"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		target := b.Get([]byte(slug))
		if target == nil {
			return fmt.Errorf("no value for target in %s", "__contentIndex")
		}
		j, err := Upload(string(target))
		if err != nil {
			return err
		}

		_, err = val.Write(j)

		return err
	})

	return val.Bytes(), err
}

// func replaceUpload(id string, data url.Values) error {
// 	// Unmarsal the existing values
// 	s := t()
// 	err := json.Unmarshal(existingContent, &s)
// 	if err != nil {
// 		log.Println("Error decoding json while updating", ns, ":", err)
// 		return j, err
// 	}

// 	// Don't allow the Item fields to be updated from form values
// 	data.Del("id")
// 	data.Del("uuid")
// 	data.Del("slug")

// 	dec := schema.NewDecoder()
// 	dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
// 	dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
// 	err = dec.Decode(s, data)
// 	if err != nil {
// 		return j, err
// 	}

// 	j, err = json.Marshal(s)
// 	if err != nil {
// 		return j, err
// 	}

// 	return j, nil
// 	return nil
// }
