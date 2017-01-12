package item

import "errors"

const (
	typeNotRegistered = `Error:
There is no type registered for %[1]s

Add this to the file which defines %[1]s{} in the 'content' package:


	func init() {			
		item.Types["%[1]s"] = func() interface{} { return new(%[1]s) }
	}		
				

`
)

var (
	// ErrTypeNotRegistered means content type isn't registered (not found in Types map)
	ErrTypeNotRegistered = errors.New(typeNotRegistered)

	// ErrAllowHiddenItem should be used as an error to tell a caller of Hideable#Hide
	// that this type is hidden, but should be shown in a particular case, i.e.
	// if requested by a valid admin or user
	ErrAllowHiddenItem = errors.New(`Allow hidden item`)

	// Types is a map used to reference a type name to its actual Editable type
	// mainly for lookups in /admin route based utilities
	Types map[string]func() interface{}
)

func init() {
	Types = make(map[string]func() interface{})
}
