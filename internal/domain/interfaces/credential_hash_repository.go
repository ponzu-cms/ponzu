package interfaces

import "github.com/fanky5g/ponzu/internal/domain/entities"

type CredentialHashRepositoryInterface interface {
	GetByUserId(userId string, credentialType entities.CredentialType) (*entities.CredentialHash, error)
	SetCredential(hash *entities.CredentialHash) error
}
