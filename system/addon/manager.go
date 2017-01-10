package addon

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"

	"github.com/gorilla/schema"
	"github.com/ponzu-cms/ponzu/management/editor"
)

const defaultInput = `<input type="hidden" name="%s" value="%s"/>`

const managerHTML = `
<div class="card editor">
    <form method="post" action="/admin/addon" enctype="multipart/form-data">
		{{ .DefaultInputs }}
		{{ .Editor }}
		<div class="row">
			<button type="submit" class="btn green waves-effect waves-light right">Save</button>
		</div>
	</form>
</div>
`

type manager struct {
	DefaultInputs template.HTML
	Editor        template.HTML
}

// Manage ...
func Manage(data url.Values, reverseDNS string) ([]byte, error) {
	a, ok := Types[reverseDNS]
	if !ok {
		return nil, fmt.Errorf("Addon has not been added to addon.Types map")
	}

	at := a()

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err := dec.Decode(at, data)
	if err != nil {
		return nil, err
	}

	e, ok := at.(editor.Editable)
	if !ok {
		return nil, fmt.Errorf("Addon is not editable - must implement editor.Editable: %T", at)
	}

	v, err := e.MarshalEditor()
	if err != nil {
		return nil, fmt.Errorf("Couldn't marshal editor for addon: %s", err.Error())
	}

	inputs := &bytes.Buffer{}
	fields := []string{
		"addon_name",
		"addon_author",
		"addon_author_url",
		"addon_version",
		"addon_reverse_dns",
		"addon_status",
	}

	for _, f := range fields {
		input := fmt.Sprintf(defaultInput, f, data.Get(f))
		_, err := inputs.WriteString(input)
		if err != nil {
			return nil, fmt.Errorf("Failed to write input for addon view: %s", f)
		}
	}

	m := manager{
		DefaultInputs: template.HTML(inputs.Bytes()),
		Editor:        template.HTML(v),
	}

	// execute html template into buffer for func return val
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("manager").Parse(managerHTML))
	tmpl.Execute(buf, m)

	return buf.Bytes(), nil
}
