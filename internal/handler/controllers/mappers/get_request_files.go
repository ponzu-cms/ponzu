package mappers

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"mime/multipart"
	"net/http"
)

func GetRequestFiles(req *http.Request) (map[string]*multipart.FileHeader, error) {
	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		return nil, err
	}

	files := make(map[string]*multipart.FileHeader)
	for fieldName, fds := range req.MultipartForm.File {
		for _, f := range fds {
			var filename string
			filename, err = item.NormalizeString(f.Filename)
			if err != nil {
				return nil, err
			}

			files[fmt.Sprintf("%s:%s", fieldName, filename)] = f
		}
	}

	return files, nil
}
