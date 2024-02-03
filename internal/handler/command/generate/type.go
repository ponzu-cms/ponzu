package generate

import (
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"strings"
)

// blog title:string Author:string PostCategory:string content:string some_thing:int
func parseType(args []string) (*entities.TypeDefinition, error) {
	t := &entities.TypeDefinition{
		Name: fieldName(args[0]),
	}
	t.Initial = strings.ToLower(string(t.Name[0]))

	fields := args[1:]
	for _, field := range fields {
		f, err := parseField(field, t)
		if err != nil {
			return nil, err
		}

		// set initial (1st character of the type's name) on field, so we don't need
		// to set the template variable like was done in prior version
		f.Initial = t.Initial
		t.Fields = append(t.Fields, *f)
	}

	return t, nil
}
