package contentgenerator

import (
	"bytes"
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/entities"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

type generator struct {
	templateDir string
	contentDir  string
}

func (gt *generator) Generate(definition *entities.TypeDefinition) error {
	for i := range definition.Fields {
		if err := gt.setFieldView(&definition.Fields[i]); err != nil {
			return err
		}
	}

	fileName := strings.ToLower(definition.Name) + ".go"
	filePath := filepath.Join(gt.contentDir, fileName)

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		localFile := filepath.Join(gt.contentDir, fileName)
		return fmt.Errorf("please remove '%s' before executing this command", localFile)
	}

	tmplPath := filepath.Join(gt.templateDir, "gen-content.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %s", err.Error())
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, definition)
	if err != nil {
		return fmt.Errorf("failed to execute template: %s", err.Error())
	}

	fmtBuf, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format template: %s", err.Error())
	}

	// no file exists. ok to write new one
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Printf("Failed to close file: %v", err)
		}
	}(file)

	_, err = file.Write(fmtBuf)
	if err != nil {
		return fmt.Errorf("failed to write generated file buffer: %s", err.Error())
	}

	return nil
}

func New() (interfaces.ContentGenerator, error) {
	_, b, _, _ := runtime.Caller(0)
	rootPath := filepath.Join(filepath.Dir(b), "../../../..")
	templateDir := filepath.Join(rootPath, "internal", "domain", "services", "contentgenerator", "templates")
	contentDir := filepath.Join(rootPath, "internal", "domain", "entities", "content")

	return &generator{
		templateDir: templateDir,
		contentDir:  contentDir,
	}, nil
}
