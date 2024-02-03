package generate

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"strings"
)

func parseField(raw string, gt *entities.TypeDefinition) (*entities.Field, error) {
	// contents:string
	// contents:string:richtext
	// author:@author,name,age
	// authors:[]@author,name,age

	if !strings.Contains(raw, ":") {
		return nil, fmt.Errorf("invalid generate argument. [%s]", raw)
	}

	data := strings.Split(raw, ":")

	field := &entities.Field{
		Name:     fieldName(data[0]),
		Initial:  gt.Initial,
		JSONName: fieldJSONName(data[0]),
	}

	setFieldTypeName(field, data[1], gt)
	viewType := "input"
	if len(data) == 3 {
		viewType = data[2]
	}

	field.ViewType = viewType
	return field, nil
}

// parse the field's type name and check if it is a special reference type, or
// a slice of reference types, which we'll set their underlying type to string
// or []string respectively
func setFieldTypeName(field *entities.Field, fieldType string, gt *entities.TypeDefinition) {
	if !strings.Contains(fieldType, "@") {
		// not a reference, set as-is downcased
		field.TypeName = strings.ToLower(fieldType)
		field.IsReference = false
		return
	}

	// some possibilities are
	// @author,name,age
	// []@author,name,age
	// -------------------
	// [] = slice of author
	// @author = reference to Author struct
	// ,name,age = JSON tag names from Author struct fields to use as select option display

	if strings.Contains(fieldType, ",") {
		referenceConf := strings.Split(fieldType, ",")
		fieldType = referenceConf[0]
		field.ReferenceJSONTags = referenceConf[1:]
	}

	var referenceType string
	if strings.HasPrefix(fieldType, "[]") {
		referenceType = strings.TrimPrefix(fieldType, "[]@")
		fieldType = "[]string"
	} else {
		referenceType = strings.TrimPrefix(fieldType, "@")
		fieldType = "string"
	}

	field.TypeName = strings.ToLower(fieldType)
	field.ReferenceName = fieldName(referenceType)
	field.IsReference = true
	gt.HasReferences = true
	return
}

// get the initial field name passed and check it for all possible cases
// MyTitle:string myTitle:string my_title:string -> MyTitle
// error-message:string -> ErrorMessage
func fieldName(name string) string {
	// remove _ or - if first character
	if name[0] == '-' || name[0] == '_' {
		name = name[1:]
	}

	// remove _ or - if last character
	if name[len(name)-1] == '-' || name[len(name)-1] == '_' {
		name = name[:len(name)-1]
	}

	// upcase the first character
	name = strings.ToUpper(string(name[0])) + name[1:]

	// remove _ or - character, and upcase the character immediately following
	for i := 0; i < len(name); i++ {
		r := rune(name[i])
		if isUnderscore(r) || isHyphen(r) {
			up := strings.ToUpper(string(name[i+1]))
			name = name[:i] + up + name[i+2:]
		}
	}

	return name
}

// get the initial field name passed and convert to json-like name
// MyTitle:string myTitle:string my_title:string -> my_title
// error-message:string -> error-message
func fieldJSONName(name string) string {
	// remove _ or - if first character
	if name[0] == '-' || name[0] == '_' {
		name = name[1:]
	}

	// downcase the first character
	name = strings.ToLower(string(name[0])) + name[1:]

	// check for uppercase character, downcase and insert _ before it if i-1
	// isn't already _ or -
	for i := 0; i < len(name); i++ {
		r := rune(name[i])
		if isUpper(r) {
			low := strings.ToLower(string(r))
			if name[i-1] == '_' || name[i-1] == '-' {
				name = name[:i] + low + name[i+1:]
			} else {
				name = name[:i] + "_" + low + name[i+1:]
			}
		}
	}

	return name
}

func isUpper(char rune) bool {
	if char >= 'A' && char <= 'Z' {
		return true
	}

	return false
}

func isUnderscore(char rune) bool {
	return char == '_'
}

func isHyphen(char rune) bool {
	return char == '-'
}
