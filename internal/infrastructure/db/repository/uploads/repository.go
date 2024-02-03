package uploads

import (
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"github.com/fanky5g/ponzu/internal/infrastructure/db/repository/root"
)

func New(db *bolt.DB) (interfaces.ContentRepositoryInterface, error) {
	return root.New(db, map[string]item.EntityBuilder{
		"uploads": func() interface{} {
			return new(item.FileUpload)
		},
	})
}