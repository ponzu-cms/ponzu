package interfaces

import (
	"github.com/fanky5g/ponzu/internal/domain/entities"
)

type ContentGenerator interface {
	Generate(typeDefinition *entities.TypeDefinition) error
	ValidateField(field *entities.Field) error
}
