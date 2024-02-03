package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func (s *service) LoginByEmail(email string, credential *entities.Credential) (string, time.Time, error) {
	var tokenExpires time.Time
	user, err := s.userRepository.GetUserByEmail(email)
	if err != nil {
		return "", tokenExpires, fmt.Errorf("failed to get user by email: %v", err)
	}

	if user == nil {
		return "", tokenExpires, errors.New("invalid user")
	}

	if err = s.VerifyCredential(user.ID, credential); err != nil {
		return "", tokenExpires, errors.New("invalid credentials")
	}

	return s.NewToken(user)
}

// checkPassword compares the hash with the salted password. A nil return means
// the password is correct, but an error could mean either the password is not
// correct, or the salt process failed - indicated in logs
func checkPassword(salt, hash, password string) error {
	stdDecodedSalt, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return fmt.Errorf("failed to decode user salt: %v", err)
	}

	salted, err := saltPassword([]byte(password), stdDecodedSalt)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), salted)
}
