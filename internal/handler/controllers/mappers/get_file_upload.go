package mappers

import (
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"net/url"
)

func GetFileUploadFromFormData(data url.Values) (interface{}, error) {
	return getGenericEntity(func() interface{} {
		return new(item.FileUpload)
	}, data)
}
