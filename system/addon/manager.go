package addon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"

	"github.com/ponzu-cms/ponzu/management/editor"

	"github.com/gorilla/schema"
	"github.com/tidwall/gjson"
)

const defaultInput = `<input type="hidden" name="%s" value="%s"/>`

const managerHTML = `
<div class="card editor">
    <form method="post" action="/admin/addon" enctype="multipart/form-data">
		<div class="card-content">
			<div class="card-title">{{ .AddonName }}</div>
		</div>
		{{ .DefaultInputs }}
		{{ .Editor }}
	</form>
</div>
`

type manager struct {
	DefaultInputs template.HTML
	Editor        template.HTML
	AddonName     string
}

// Manage ...
func Manage(data []byte, reverseDNS string) ([]byte, error) {
	a, ok := Types[reverseDNS]
	if !ok {
		return nil, fmt.Errorf("Addon has not been added to addon.Types map")
	}

	// convert json => map[string]interface{} => url.Values
	var kv map[string]interface{}
	err := json.Unmarshal(data, &kv)
	if err != nil {
		return nil, err
	}

	vals := make(url.Values)
	for k, v := range kv {
		switch v.(type) {
		case []string:
			s := v.([]string)
			for i := range s {
				if i == 0 {
					vals.Set(k, s[i])
				}

				vals.Add(k, s[i])
			}
		default:
			vals.Set(k, fmt.Sprintf("%v", v))
		}
	}

	at := a()

	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err = dec.Decode(at, vals)
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
		input := fmt.Sprintf(defaultInput, f, gjson.GetBytes(data, f).String())
		_, err := inputs.WriteString(input)
		if err != nil {
			return nil, fmt.Errorf("Failed to write input for addon view: %s", f)
		}
	}

	m := manager{
		DefaultInputs: template.HTML(inputs.Bytes()),
		Editor:        template.HTML(v),
		AddonName:     gjson.GetBytes(data, "addon_name").String(),
	}

	// execute html template into buffer for func return val
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("manager").Parse(managerHTML))
	tmpl.Execute(buf, m)

	return buf.Bytes(), nil
}
