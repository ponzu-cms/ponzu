// Package user contains the basic admin user creation and authentication code,
// specific to Ponzu systems.
package user

import (
	"bytes"
	crand "crypto/rand"
	"encoding/base64"
	"log"
	mrand "math/rand"
	"net/http"
	"time"

	"github.com/nilslice/jwt"
	"golang.org/x/crypto/bcrypt"
)

// User defines a admin user in the system
type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Hash  string `json:"hash"`
	Salt  string `json:"salt"`
}

var (
	r = mrand.New(mrand.NewSource(time.Now().Unix()))
)

// New creates a user
func New(email, password string) (*User, error) {
	salt, err := randSalt()
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword([]byte(password), salt)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email: email,
		Hash:  string(hash),
		Salt:  base64.StdEncoding.EncodeToString(salt),
	}

	return user, nil
}

// Auth is HTTP middleware to ensure the request has proper token credentials
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		redir := req.URL.Scheme + req.URL.Host + "/admin/login"

		if IsValid(req) {
			next.ServeHTTP(res, req)
		} else {
			http.Redirect(res, req, redir, http.StatusFound)
		}
	})
}

// IsValid checks if the user request is authenticated
func IsValid(req *http.Request) bool {
	// check if token exists in cookie
	cookie, err := req.Cookie("_token")
	if err != nil {
		return false
	}
	// validate it and allow or redirect request
	token := cookie.Value
	return jwt.Passes(token)
}

// IsUser checks for consistency in email/pass combination
func IsUser(usr *User, password string) bool {
	salt, err := base64.StdEncoding.DecodeString(usr.Salt)
	if err != nil {
		return false
	}

	err = checkPassword([]byte(usr.Hash), []byte(password), salt)
	if err != nil {
		log.Println("Error checking password:", err)
		return false
	}

	return true
}

// randSalt generates 16 * 8 bits of data for a random salt
func randSalt() ([]byte, error) {
	buf := make([]byte, 16)
	count := len(buf)
	n, err := crand.Read(buf)
	if err != nil {
		return nil, err
	}

	if n != count || err != nil {
		for count > 0 {
			count--
			buf[count] = byte(r.Int31n(256))
		}
	}

	return buf, nil
}

// saltPassword combines the salt and password provided
func saltPassword(password, salt []byte) ([]byte, error) {
	salted := &bytes.Buffer{}
	_, err := salted.Write(append(salt, password...))
	if err != nil {
		return nil, err
	}

	return salted.Bytes(), nil
}

// hashPassword encrypts the salted password using bcrypt
func hashPassword(password, salt []byte) ([]byte, error) {
	salted, err := saltPassword(password, salt)
	if err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword(salted, 10)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// checkPassword compares the hash with the salted password. A nil return means
// the password is correct, but an error could mean either the password is not
// correct, or the salt process failed - indicated in logs
func checkPassword(hash, password, salt []byte) error {
	salted, err := saltPassword(password, salt)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword(hash, salted)
}
