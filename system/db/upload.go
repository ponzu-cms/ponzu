package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ponzu-cms/ponzu/system/item"

	"github.com/boltdb/bolt"
	"github.com/gorilla/schema"
	uuid "github.com/satori/go.uuid"
)

// SetUpload stores information about files uploaded to the system
func SetUpload(data url.Values) error {
	// set new UUID for upload
	data.Set("uuid", uuid.NewV4().String())

	// create slug based on filename and timestamp/updated fields
	slug := data.Get("name")
	slug, err := checkSlugForDuplicate(slug)
	if err != nil {
		return err
	}
	data.Set("slug", slug)

	ts := fmt.Sprintf("%d", time.Now().Unix()*1000)
	data.Set("timestamp", ts)
	data.Set("updated", ts)

	// store in database
	err = store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("__uploads"))
		if err != nil {
			return err
		}

		// get sequential ID for item
		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		data.Set("id", fmt.Sprintf("%d", id))

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
		// -

		return nil
	})

	return err
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
