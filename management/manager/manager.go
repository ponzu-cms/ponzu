package manager

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/nilslice/cms/management/editor"
)

var html = `
<a href="/admin/edit?type={{.Kind}}" class="button">New {{.Kind}}</a>
<div class="manager">
    <form method="post" action="/admin/edit">
        {{.Editor}}
		<input type="hidden" name="id" value="{{.ID}}"/>
		<input type="hidden" name="type" value="{{.Kind}}"/>
        <input type="submit" value="Save"/>
    </form>
</div>
`

type form struct {
	ID     int
	Kind   string
	Editor template.HTML
}

// Manage ...
func Manage(e editor.Editable, typeName string) ([]byte, error) {
	v, err := e.MarshalEditor()
	if err != nil {
		return nil, fmt.Errorf("Couldn't marshal editor for content %T. %s", e, err.Error())
	}

	f := form{
		ID:     e.ContentID(),
		Kind:   typeName,
		Editor: template.HTML(v),
	}

	// execute html template into buffer for func return val
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("manager").Parse(html))
	tmpl.Execute(buf, f)

	return buf.Bytes(), nil
}
