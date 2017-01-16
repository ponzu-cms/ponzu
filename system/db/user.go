package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/ponzu-cms/ponzu/system/admin/user"

	"github.com/boltdb/bolt"
	"github.com/nilslice/jwt"
)

// ErrUserExists is used for the db to report to admin user of existing user
var ErrUserExists = errors.New("Error. User exists.")

// ErrNoUserExists is used for the db to report to admin user of non-existing user
var ErrNoUserExists = errors.New("Error. No user exists.")

// SetUser sets key:value pairs in the db for user settings
func SetUser(usr *user.User) (int, error) {
	err := store.Update(func(tx *bolt.Tx) error {
		email := []byte(usr.Email)
		users := tx.Bucket([]byte("__users"))
		if users == nil {
			return bolt.ErrBucketNotFound
		}

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

		err = users.Put(email, j)
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

// UpdateUser sets key:value pairs in the db for existing user settings
func UpdateUser(usr, updatedUsr *user.User) error {
	// ensure user ID remains the same
	if updatedUsr.ID != usr.ID {
		updatedUsr.ID = usr.ID
	}

	err := store.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte("__users"))
		if users == nil {
			return bolt.ErrBucketNotFound
		}

		// check if user is found by email, fail if nil
		exists := users.Get([]byte(usr.Email))
		if exists == nil {
			return ErrNoUserExists
		}

		// marshal User to json and put into bucket
		j, err := json.Marshal(updatedUsr)
		if err != nil {
			return err
		}

		err = users.Put([]byte(updatedUsr.Email), j)
		if err != nil {
			return err
		}

		// if email address was changed, delete the old record of former
		// user with original email address
		if usr.Email != updatedUsr.Email {
			err = users.Delete([]byte(usr.Email))
			if err != nil {
				return err
			}

		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes a user from the db by email
func DeleteUser(email string) error {
	err := store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__users"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		err := b.Delete([]byte(email))
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

// User gets the user by email from the db
func User(email string) ([]byte, error) {
	val := &bytes.Buffer{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__users"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

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

	if val.Bytes() == nil {
		return nil, ErrNoUserExists
	}

	return val.Bytes(), nil
}

// UserAll returns all users from the db
func UserAll() ([][]byte, error) {
	var users [][]byte
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__users"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		err := b.ForEach(func(k, v []byte) error {
			users = append(users, v)
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}

// CurrentUser extracts the user from the request data and returns the current user from the db
func CurrentUser(req *http.Request) ([]byte, error) {
	if !user.IsValid(req) {
		return nil, fmt.Errorf("Error. Invalid User.")
	}

	token, err := req.Cookie("_token")
	if err != nil {
		return nil, err
	}

	claims := jwt.GetClaims(token.Value)
	email, ok := claims["user"]
	if !ok {
		return nil, fmt.Errorf("Error. No user data found in request token.")
	}

	usr, err := User(email.(string))
	if err != nil {
		return nil, err
	}

	return usr, nil
}

// SetRecoveryKey generates and saves a random secret key to verify an email
// address submitted in order to recover/reset an account password
func SetRecoveryKey(email string) (string, error) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	key := fmt.Sprintf("%d", r.Int63())

	err := store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("__recoveryKeys"))
		if err != nil {
			return err
		}

		err = b.Put([]byte(email), []byte(key))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return key, nil
}

// RecoveryKey gets a previously set recovery key to verify an email address
// submitted in order to recover/reset an account password
func RecoveryKey(email string) (string, error) {
	key := &bytes.Buffer{}

	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__recoveryKeys"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		_, err := key.Write(b.Get([]byte(email)))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return key.String(), nil
}
