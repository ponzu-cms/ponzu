package content

import "net/http"

const (
	// ErrTypeNotRegistered means content type isn't registered (not found in Types map)
	ErrTypeNotRegistered = `Error:
There is no type registered for %[1]s

Add this to the file which defines %[1]s{} in the 'content' package:
--------------------------------------------------------------------------+

func init() {			
	Types["%[1]s"] = func() interface{} { return new(%[1]s) }
}		
				
--------------------------------------------------------------------------+
`
)

// Mergeable allows external post content to be approved and published through
// the public-facing API
type Mergeable interface {
	// Approve copies an external post to the internal collection and triggers
	// a re-sort of its content type posts
	Approve(req *http.Request) error
}

// Types is a map used to reference a type name to its actual Editable type
// mainly for lookups in /admin route based utilities
var Types = make(map[string]func() interface{})
