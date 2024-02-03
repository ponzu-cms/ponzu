package views

import (
	"bytes"
	"github.com/fanky5g/ponzu/internal/util"
)

// Login ...
func Login(appName string) ([]byte, error) {
	a := View{
		Logo: appName,
	}

	buf := &bytes.Buffer{}
	tmpl := util.MakeTemplate("start_admin", "login_admin", "end_admin")
	err := tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
