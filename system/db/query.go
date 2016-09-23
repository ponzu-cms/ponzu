package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/gorilla/schema"
	"github.com/nilslice/cms/content"
	"github.com/nilslice/cms/management/editor"
	"github.com/nilslice/cms/management/manager"
)

var store *bolt.DB

func init() {
	var err error
	store, err = bolt.Open("store.db", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}

	// initialize db with all content type buckets
	store.Update(func(tx *bolt.Tx) error {
		for t := range content.Types {
			_, err := tx.CreateBucketIfNotExists([]byte(t))
			if err != nil {
				return err
			}
		}

		return nil
	})

}

// Set inserts or updates values in the database.
// The `target` argument is a string made up of namespace:id (string:int)
func Set(target string, data url.Values) (int, error) {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	// check if content id == -1 (indicating new post).
	// if so, run an insert which will assign the next auto incremented int.
	// this is done because boltdb begins its bucket auto increment value at 0,
	// which is the zero-value of an int in the Item struct field for ID.
	// this is a problem when the original first post (with auto ID = 0) gets
	// overwritten by any new post, originally having no ID, defauting to 0.
	if id == "-1" {
		return insert(ns, data)
	}

	return update(ns, id, data)
}

func update(ns, id string, data url.Values) (int, error) {
	cid, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}

	err = store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(ns))
		if err != nil {
			return err
		}

		j, err := toJSON(ns, data)
		if err != nil {
			return err
		}

		err = b.Put([]byte(fmt.Sprintf("%d", cid)), j)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, nil
	}

	return cid, nil
}

func insert(ns string, data url.Values) (int, error) {
	var effectedID int
	err := store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(ns))
		if err != nil {
			return err
		}

		// get the next available ID and convert to string
		// also set effectedID to int of ID
		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		cid := strconv.FormatUint(id, 10)
		effectedID, err = strconv.Atoi(cid)
		if err != nil {
			return err
		}
		data.Add("id", cid)

		j, err := toJSON(ns, data)
		if err != nil {
			return err
		}

		err = b.Put([]byte(cid), j)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return effectedID, nil
}

func toJSON(ns string, data url.Values) ([]byte, error) {
	// find the content type and decode values into it
	t, ok := content.Types[ns]
	if !ok {
		return nil, fmt.Errorf(content.ErrTypeNotRegistered, ns)
	}
	post := t()

	dec := schema.NewDecoder()
	dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
	dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
	err := dec.Decode(post, data)
	if err != nil {
		return nil, err
	}

	slug, err := manager.Slug(post.(editor.Editable))
	if err != nil {
		return nil, err
	}
	post.(editor.Editable).SetSlug(slug)

	// marshall content struct to json for db storage
	j, err := json.Marshal(post)
	if err != nil {
		return nil, err
	}

	return j, nil
}

// Get retrives one item from the database. Non-existent values will return an empty []byte
// The `target` argument is a string made up of namespace:id (string:int)
func Get(target string) ([]byte, error) {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	val := &bytes.Buffer{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		_, err := val.Write(b.Get([]byte(id)))
		if err != nil {
			fmt.Println(err)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return val.Bytes(), nil
}

// GetAll retrives all items from the database within the provided namespace
func GetAll(namespace string) [][]byte {
	var posts [][]byte
	store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(namespace))

		len := b.Stats().KeyN
		posts = make([][]byte, 0, len)

		b.ForEach(func(k, v []byte) error {
			posts = append(posts, v)

			return nil
		})

		return nil
	})

	return posts
}
