package user

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/nilslice/jwt"
	"github.com/nilslice/rand"
	"golang.org/x/crypto/bcrypt"
)

// User defines a admin user in the system
type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Hash  string `json:"hash"`
	Salt  string `json:"salt"`
}

// NewUser creates a user
func NewUser(email, password string) *User {
	salt := salt128()
	hash := encryptPassword([]byte(password), salt)

	user := &User{
		Email: email,
		Hash:  string(hash),
		Salt:  base64.StdEncoding.EncodeToString(salt),
	}

	return user
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
	fmt.Println(usr, password)
	salt, err := base64.StdEncoding.DecodeString(usr.Salt)
	if err != nil {
		return false
	}

	err = comparePassword([]byte(usr.Hash), []byte(password), salt)
	if err != nil {
		return false
	}

	return true
}

// The following functions are from github.com/sluu99/um -----------------------

// salt128 generates 128 bits of random data.
func salt128() []byte {
	x := make([]byte, 16)
	rand.Read(x)
	return x
}

// makePassword makes the actual password from the plain password and salt
func makePassword(plainPw, salt []byte) []byte {
	password := make([]byte, 0, len(plainPw)+len(salt))
	password = append(password, salt...)
	password = append(password, plainPw...)
	return password
}

// encryptPassword uses bcrypt to encrypt a password and salt combination.
// It returns the encrypted password in hex form.
func encryptPassword(plainPw, salt []byte) []byte {
	hash, _ := bcrypt.GenerateFromPassword(makePassword(plainPw, salt), 10)
	return hash
}

// comparePassword compares a hash with the plain password and the salt.
// The function returns nil on success or an error on failure.
func comparePassword(hash, plainPw, salt []byte) error {
	return bcrypt.CompareHashAndPassword(hash, makePassword(plainPw, salt))
}

// End code from github.com/sluu99/um ------------------------------------------
