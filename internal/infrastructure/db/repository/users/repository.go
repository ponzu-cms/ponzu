package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	domainErrors "github.com/fanky5g/ponzu/internal/domain/errors"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
)

var (
	bucketName = "__users"
)

type repository struct {
	db *bolt.DB
}

// SetUser sets key:value pairs in the db for user settings
func (repo *repository) SetUser(usr *entities.User) error {
	return repo.db.Update(func(tx *bolt.Tx) error {
		email := []byte(usr.Email)
		users := tx.Bucket([]byte(bucketName))
		if users == nil {
			return bolt.ErrBucketNotFound
		}

		// check if user is found by email, fail if nil
		exists := users.Get(email)
		if exists != nil {
			return domainErrors.ErrUserExists
		}

		// get NextSequence int64 and set it as the GetUserByEmail.ID
		id, err := users.NextSequence()
		if err != nil {
			return err
		}

		usr.ID = fmt.Sprint(id)

		// marshal GetUserByEmail to json and put into bucket
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
}

// UpdateUser sets key:value pairs in the db for existing user settings
func (repo *repository) UpdateUser(usr, updatedUsr *entities.User) error {
	// ensure user ID remains the same
	if updatedUsr.ID != usr.ID {
		updatedUsr.ID = usr.ID
	}

	return repo.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte(bucketName))
		if users == nil {
			return bolt.ErrBucketNotFound
		}

		// check if user is found by email, fail if nil
		exists := users.Get([]byte(usr.Email))
		if exists == nil {
			return domainErrors.ErrNoUserExists
		}

		// marshal GetUserByEmail to json and put into bucket
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
}

// DeleteUser deletes a user from the db by email
func (repo *repository) DeleteUser(email string) error {
	err := repo.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
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

// GetUserByEmail gets the user by email from the db
func (repo *repository) GetUserByEmail(email string) (*entities.User, error) {
	val := &bytes.Buffer{}
	err := repo.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
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

	userBytes := val.Bytes()
	if userBytes == nil {
		return nil, domainErrors.ErrNoUserExists
	}

	var user entities.User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAllUsers returns all users from the db
func (repo *repository) GetAllUsers() ([][]byte, error) {
	var users [][]byte
	err := repo.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
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

func New(db *bolt.DB) (interfaces.UserRepositoryInterface, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to create storage bucket: %v", bucketName)
	}

	return &repository{db: db}, nil
}
