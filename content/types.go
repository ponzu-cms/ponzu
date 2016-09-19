package content

import "github.com/nilslice/cms/management/editor"

// Types is a map used to reference a type name to its actual Editable type
// mainly for lookups in /admin route based utilities
var Types = make(map[string]editor.Editable)
