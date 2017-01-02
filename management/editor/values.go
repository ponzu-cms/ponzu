package editor

import (
	"fmt"
	"reflect"
	"strings"
)

// TagNameFromStructField does a lookup on the `json` struct tag for a given
// field of a struct
func TagNameFromStructField(name string, post interface{}) string {
	// sometimes elements in these environments will not have a name,
	// and thus no tag name in the struct which correlates to it.
	if name == "" {
		return name
	}

	field, ok := reflect.TypeOf(post).Elem().FieldByName(name)
	if !ok {
		panic("Couldn't get struct field for: " + name + ". Make sure you pass the right field name to editor field elements.")
	}

	tag, ok := field.Tag.Lookup("json")
	if !ok {
		panic("Couldn't get json struct tag for: " + name + ". Struct fields for content types must have 'json' tags.")
	}

	return tag
}

// TagNameFromStructFieldMulti calls TagNameFromStructField and formats is for
// use with gorilla/schema
// due to the format in which gorilla/schema expects form names to be when
// one is associated with multiple values, we need to output the name as such.
// Ex. 'category.0', 'category.1', 'category.2' and so on.
func TagNameFromStructFieldMulti(name string, i int, post interface{}) string {
	tag := TagNameFromStructField(name, post)

	return fmt.Sprintf("%s.%d", tag, i)
}

// ValueFromStructField returns the string value of a field in a struct
func ValueFromStructField(name string, post interface{}) string {
	field := reflect.Indirect(reflect.ValueOf(post)).FieldByName(name)

	switch field.Kind() {
	case reflect.String:
		return field.String()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%v", field.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%v", field.Uint())

	case reflect.Bool:
		return fmt.Sprintf("%t", field.Bool())

	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%v", field.Complex())

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", field.Float())

	case reflect.Slice:
		s := []string{}

		for i := 0; i < field.Len(); i++ {
			pos := field.Index(i)
			s = append(s, fmt.Sprintf("%v", pos))
		}

		return strings.Join(s, "__ponzu")

	default:
		panic(fmt.Sprintf("Ponzu: Type '%s' for field '%s' not supported.", field.Type(), name))
	}
}
