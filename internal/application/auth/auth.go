package auth

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/application"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"github.com/nilslice/jwt"
	"math/rand"
	"time"
)

var ServiceToken application.ServiceToken = "AuthService"

type service struct {
	userRepository        interfaces.UserRepositoryInterface
	credentialRepository  interfaces.CredentialHashRepositoryInterface
	recoveryKeyRepository interfaces.RecoveryKeyRepositoryInterface
}

type Service interface {
	IsTokenValid(token string) (bool, error)
	GetUserFromAuthToken(token string) (*entities.User, error)
	NewToken(user *entities.User) (string, time.Time, error)
	SetCredential(userId string, credential *entities.Credential) error
	VerifyCredential(userId string, credential *entities.Credential) error
	LoginByEmail(email string, credential *entities.Credential) (string, time.Time, error)
	GetRecoveryKey(email string) (string, error)
	SetRecoveryKey(email string) (string, error)
}

func (s *service) IsTokenValid(token string) (bool, error) {
	return jwt.Passes(token), nil
}

func (s *service) GetUserFromAuthToken(token string) (*entities.User, error) {
	isValid, err := s.IsTokenValid(token)
	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, fmt.Errorf("error. Invalid token")
	}

	claims := jwt.GetClaims(token)
	email, ok := claims["user"]
	if !ok {
		return nil, fmt.Errorf("error. No user data found in request token")
	}

	return s.userRepository.GetUserByEmail(email.(string))
}

func (s *service) NewToken(user *entities.User) (string, time.Time, error) {
	// create new token
	expires := time.Now().Add(time.Hour * 24 * 7)
	claims := map[string]interface{}{
		"exp":  expires,
		"user": user.Email,
	}

	token, err := jwt.New(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expires, nil
}

func (s *service) GetRecoveryKey(email string) (string, error) {
	return s.recoveryKeyRepository.GetRecoveryKey(email)
}

func (s *service) SetRecoveryKey(email string) (string, error) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	key := fmt.Sprintf("%d", r.Int63())

	return key, s.recoveryKeyRepository.SetRecoveryKey(email, key)
}

func New(
	configRepository interfaces.ConfigRepositoryInterface,
	userRepository interfaces.UserRepositoryInterface,
	credentialRepository interfaces.CredentialHashRepositoryInterface,
	recoveryKeyRepository interfaces.RecoveryKeyRepositoryInterface) (Service, error) {
	clientSecret := configRepository.Cache().GetByKey("client_secret").(string)
	if clientSecret != "" {
		jwt.Secret([]byte(clientSecret))
	}

	return &service{
		userRepository:        userRepository,
		credentialRepository:  credentialRepository,
		recoveryKeyRepository: recoveryKeyRepository,
	}, nil
}
