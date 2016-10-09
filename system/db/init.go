package db

import (
	"encoding/json"
	"log"

	"../../content"
	"../../system/admin/config"

	"github.com/boltdb/bolt"
	"github.com/nilslice/jwt"
)

var store *bolt.DB

// Init creates a db connection, initializes db with required info, sets secrets
func Init() {
	var err error
	store, err = bolt.Open("store.db", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = store.Update(func(tx *bolt.Tx) error {
		// initialize db with all content type buckets
		for t := range content.Types {
			_, err := tx.CreateBucketIfNotExists([]byte(t))
			if err != nil {
				return err
			}
		}

		// init db with other buckets as needed
		buckets := []string{"_config", "_users"}
		for _, name := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return err
			}
		}

		// seed db with configs structure if not present
		b := tx.Bucket([]byte("_config"))
		if b.Get([]byte("settings")) == nil {
			j, err := json.Marshal(&config.Config{})
			if err != nil {
				return err
			}

			err = b.Put([]byte("settings"), j)
			if err != nil {
				return err
			}
		}

		clientSecret := ConfigCache("client_secret")

		if clientSecret != "" {
			jwt.Secret([]byte(clientSecret))
		}

		return nil
	})
	if err != nil {
		log.Fatal("Coudn't initialize db with buckets.", err)
	}

}

// SystemInitComplete checks if there is at least 1 admin user in the db which
// would indicate that the system has been configured to the minimum required.
func SystemInitComplete() bool {
	complete := false

	err := store.View(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte("_users"))

		err := users.ForEach(func(k, v []byte) error {
			complete = true
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		complete = false
		log.Fatal(err)
	}

	return complete
}
