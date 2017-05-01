package db

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ponzu-cms/ponzu/system/item"

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

	if data.Get("uuid") == "" ||
		data.Get("uuid") == (uuid.UUID{}).String() {
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
	var id uint64
	var err error
	err = store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("__uploads"))
		if err != nil {
			return err
		}

		if pid == "-1" {
			// get sequential ID for item
			id, err = b.NextSequence()
			if err != nil {
				return err
			}
			data.Set("id", fmt.Sprintf("%d", id))
		} else {
			uid, err := strconv.ParseInt(pid, 10, 64)
			if err != nil {
				return err
			}
			id = uint64(uid)
			data.Set("id", fmt.Sprintf("%d", id))
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

		uploadKey, err := key(data.Get("id"))
		if err != nil {
			return err
		}
		err = b.Put(uploadKey, j)
		if err != nil {
			return err
		}

		// add slug to __contentIndex for lookup
		b, err = tx.CreateBucketIfNotExists([]byte("__contentIndex"))
		if err != nil {
			return err
		}

		k := []byte(data.Get("slug"))
		v := []byte(fmt.Sprintf("__uploads:%d", id))

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

	id, err := key(parts[1])
	if err != nil {
		return nil, err
	}

	err = store.View(func(tx *bolt.Tx) error {
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

		v := b.Get([]byte(slug))
		if v == nil {
			return fmt.Errorf("no value for key '%s' in __contentIndex", slug)
		}

		j, err := Upload(string(v))
		if err != nil {
			return err
		}

		_, err = val.Write(j)

		return err
	})

	return val.Bytes(), err
}

// UploadAll returns a [][]byte containing all upload data from the system
func UploadAll() [][]byte {
	var uploads [][]byte
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__uploads"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		numKeys := b.Stats().KeyN
		uploads = make([][]byte, 0, numKeys)

		return b.ForEach(func(k, v []byte) error {
			uploads = append(uploads, v)
			return nil
		})
	})
	if err != nil {
		log.Println("Error in UploadAll:", err)
		return nil
	}

	return uploads
}

// DeleteUpload removes the value for an upload at its key id, based on the
// target provided i.e. __uploads:{id}
func DeleteUpload(target string) error {
	parts := strings.Split(target, ":")
	if len(parts) < 2 {
		return fmt.Errorf("Error deleting upload, invalid target %s", target)
	}
	id, err := key(parts[1])
	if err != nil {
		return err
	}

	return store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(parts[0]))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		return b.Delete(id)
	})
}

func key(sid string) ([]byte, error) {
	id, err := strconv.Atoi(sid)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return b, err
}
