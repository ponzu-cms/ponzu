package db

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/boltdb/bolt"
)

var (
	// ErrNoAddonExists indicates that there was not addon found in the db
	ErrNoAddonExists = errors.New("No addon exists.")
)

// Addon looks for an addon by its addon_reverse_dns as the key and returns
// the url.Values representation of an addon
func Addon(key string) (url.Values, error) {
	buf := &bytes.Buffer{}

	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__addons"))

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

	data, err := url.ParseQuery(buf.String())
	if err != nil {
		return nil, err
	}

	return data, nil
}

// SetAddon stores the values of an addon into the __addons bucket with a the
// addon_reverse_dns field used as the key
func SetAddon(data url.Values) error {
	// we don't know the structure of the addon type from a addon developer, so
	// encoding to json before it's stored in the db is difficult. Instead, we
	// can just encode the url.Values to a query string using the Encode() method.
	// The downside is that we will have to parse the values out of the query
	// string when loading it from the db
	v := data.Encode()

	err := store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__addons"))
		k := data.Get("addon_reverse_dns")
		if k == "" {
			name := data.Get("addon_name")
			return fmt.Errorf(`Addon "%s" has no identifier to use as key.`, name)
		}

		err := b.Put([]byte(k), []byte(v))
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
func AddonAll() []url.Values {
	var all []url.Values
	buf := &bytes.Buffer{}

	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__addons"))
		err := b.ForEach(func(k, v []byte) error {
			_, err := buf.Write(v)
			if err != nil {
				return err
			}

			data, err := url.ParseQuery(buf.String())
			if err != nil {
				return err
			}

			all = append(all, data)
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
