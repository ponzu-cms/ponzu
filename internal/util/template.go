package util

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	htmlTemplateDir string
)

func init() {
	_, b, _, _ := runtime.Caller(0)
	rootPath := filepath.Join(filepath.Dir(b), "../..")
	htmlTemplateDir = fmt.Sprintf("%s/internal/views", rootPath)
}

func MakeTemplate(templates ...string) *template.Template {
	return template.Must(template.New(strings.Join(templates, "_")).Parse(Html(templates...)))
}

func Html(templates ...string) string {
	var tmpl string
	for _, name := range templates {
		htmlString, err := getHtmlString(name)
		if err != nil {
			panic(err)
		}

		tmpl += htmlString
	}

	return tmpl
}

func getHtmlString(name string) (string, error) {
	templateName := fmt.Sprintf("%s/%s.gohtml", htmlTemplateDir, name)
	f, err := os.Open(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to open template file: %s. Error: %v", name, err)
	}

	htmlBytes, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %v", err)
	}

	return string(htmlBytes), err
}
