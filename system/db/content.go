package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/bosssauce/ponzu/content"
	"github.com/bosssauce/ponzu/management/editor"
	"github.com/bosssauce/ponzu/management/manager"
	"github.com/gorilla/schema"
)

// SetContent inserts or updates values in the database.
// The `target` argument is a string made up of namespace:id (string:int)
func SetContent(target string, data url.Values) (int, error) {
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

		j, err := postToJSON(ns, data)
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

		j, err := postToJSON(ns, data)
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

func postToJSON(ns string, data url.Values) ([]byte, error) {
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

// Content retrives one item from the database. Non-existent values will return an empty []byte
// The `target` argument is a string made up of namespace:id (string:int)
func Content(target string) ([]byte, error) {
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

// ContentAll retrives all items from the database within the provided namespace
func ContentAll(namespace string) [][]byte {
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
