package views

import (
	"bytes"
	"github.com/fanky5g/ponzu/internal/util"
)

// ForgotPassword ...
func ForgotPassword(appName string) ([]byte, error) {
	a := View{
		Logo: appName,
	}

	buf := &bytes.Buffer{}
	tmpl := util.MakeTemplate("start_admin", "forgot_password", "end_admin")
	err := tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
