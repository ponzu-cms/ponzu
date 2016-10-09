package db

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
	"github.com/bosssauce/ponzu/system/admin/user"
)

// ErrUserExists is used for the db to report to admin user of existing user
var ErrUserExists = errors.New("Error. User exists.")

// SetUser sets key:value pairs in the db for user settings
func SetUser(usr *user.User) (int, error) {
	err := store.Update(func(tx *bolt.Tx) error {
		email := []byte(usr.Email)
		users := tx.Bucket([]byte("_users"))

		// check if user is found by email, fail if nil
		exists := users.Get(email)
		if exists != nil {
			return ErrUserExists
		}

		// get NextSequence int64 and set it as the User.ID
		id, err := users.NextSequence()
		if err != nil {
			return err
		}
		usr.ID = int(id)

		// marshal User to json and put into bucket
		j, err := json.Marshal(usr)
		if err != nil {
			return err
		}

		err = users.Put([]byte(usr.Email), j)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return usr.ID, nil
}

// User gets the user by email from the db
func User(email string) ([]byte, error) {
	val := &bytes.Buffer{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("_users"))
		usr := b.Get([]byte(email))

		_, err := val.Write(usr)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return val.Bytes(), nil
}
