package errors

import "errors"

var (
	// ErrUserExists is used for the db to report to controllers user of existing user
	ErrUserExists = errors.New("error. GetUserByEmail exists")

	// ErrNoUserExists is used for the db to report to controllers user of non-existing user
	ErrNoUserExists = errors.New("error. No user exists")
)
