package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/boltdb/bolt"
	"github.com/gorilla/schema"
)

var (
	// ErrNoAddonExists indicates that there was not addon found in the db
	ErrNoAddonExists = errors.New("No addon exists.")
)

// Addon looks for an addon by its addon_reverse_dns as the key and returns
// the []byte as json representation of an addon
func Addon(key string) ([]byte, error) {
	buf := &bytes.Buffer{}

	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__addons"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		val := b.Get([]byte(key))

		if val == nil {
			return ErrNoAddonExists
		}

		_, err := buf.Write(val)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// SetAddon stores the values of an addon into the __addons bucket with a the
// `addon_reverse_dns` field used as the key. `kind` is the interface{} type for
// the provided addon (as in the result of calling addon.Types[id])
func SetAddon(data url.Values, kind interface{}) error {
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err := dec.Decode(kind, data)

	v, err := json.Marshal(kind)

	k := data.Get("addon_reverse_dns")
	if k == "" {
		name := data.Get("addon_name")
		return fmt.Errorf(`Addon "%s" has no identifier to use as key.`, name)
	}

	err = store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__addons"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		err := b.Put([]byte(k), v)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// AddonAll returns all registered addons as a [][]byte
func AddonAll() [][]byte {
	var all [][]byte

	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__addons"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		err := b.ForEach(func(k, v []byte) error {
			all = append(all, v)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Println("Error finding addons in db with db.AddonAll:", err)
		return nil
	}

	return all
}

// DeleteAddon removes an addon from the db by its key, the addon_reverse_dns
func DeleteAddon(key string) error {
	err := store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__addons"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		if err := b.Delete([]byte(key)); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// AddonExists checks if there is an existing addon stored. The key is an the
// value at addon_reverse_dns
func AddonExists(key string) bool {
	var exists bool

	if store == nil {
		Init()
	}

	err := store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("__addons"))
		if err != nil {
			return err
		}
		if b.Get([]byte(key)) == nil {
			return nil
		}

		exists = true
		return nil
	})
	if err != nil {
		log.Println("Error checking existence of addon with key:", key, "-", err)
		return false
	}

	return exists
}
