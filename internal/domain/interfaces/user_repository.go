package interfaces

import (
	"github.com/fanky5g/ponzu/internal/domain/entities"
)

type UserRepositoryInterface interface {
	SetUser(usr *entities.User) error
	UpdateUser(usr, updatedUsr *entities.User) error
	DeleteUser(email string) error
	GetUserByEmail(email string) (*entities.User, error)
	// GetAllUsers users repository can and should return Users entity and not byte arrays
	GetAllUsers() ([][]byte, error)
}
