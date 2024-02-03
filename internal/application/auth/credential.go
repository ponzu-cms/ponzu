package auth

import (
	"bytes"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"golang.org/x/crypto/bcrypt"
	mrand "math/rand"
	"time"
)

var (
	r                            = mrand.New(mrand.NewSource(time.Now().Unix()))
	ErrUnsupportedCredentialType = errors.New("unsupported credential type")
)

// SetCredential saves credential by userId. It is not responsible for checking if the user exists
func (s *service) SetCredential(userId string, credential *entities.Credential) error {
	switch credential.Type {
	case entities.CredentialTypePassword:
		salt, err := randSalt()
		if err != nil {
			return err
		}

		hash, err := hashPassword([]byte(credential.Value), salt)
		if err != nil {
			return err
		}

		passwordHash := entities.PasswordHash{
			Hash: string(hash),
			Salt: base64.StdEncoding.EncodeToString(salt),
		}

		byteValue, err := json.Marshal(passwordHash)
		if err != nil {
			return fmt.Errorf("failed to save credential: %v", err)
		}

		hashedCredential := &entities.CredentialHash{
			UserId: userId,
			Type:   entities.CredentialTypePassword,
			Value:  byteValue,
		}

		return s.credentialRepository.SetCredential(hashedCredential)
	default:
		return ErrUnsupportedCredentialType
	}
}

func (s *service) VerifyCredential(userId string, credential *entities.Credential) error {
	hashedCredential, err := s.credentialRepository.GetByUserId(userId, credential.Type)
	if err != nil {
		return err
	}

	if hashedCredential == nil {
		return errors.New("invalid credential. No match")
	}

	switch credential.Type {
	case entities.CredentialTypePassword:
		var passwordHash entities.PasswordHash
		if err = json.Unmarshal(hashedCredential.Value, &passwordHash); err != nil {
			return fmt.Errorf("failed to unmarshal credential value: %v", err)
		}

		return checkPassword(passwordHash.Salt, passwordHash.Hash, credential.Value)
	default:
		return ErrUnsupportedCredentialType
	}
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
