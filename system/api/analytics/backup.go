package analytics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
)

// Backup writes a snapshot of the analytics.db database to an HTTP response
func Backup(res http.ResponseWriter) error {
	err := store.View(func(tx *bolt.Tx) error {
		ts := time.Now().Unix()
		disposition := `attachment; filename="analytics-%d.db.bak"`

		res.Header().Set("Content-Type", "application/octet-stream")
		res.Header().Set("Content-Disposition", fmt.Sprintf(disposition, ts))
		res.Header().Set("Content-Length", fmt.Sprintf("%d", int(tx.Size())))

		_, err := tx.WriteTo(res)
		return err
	})

	return err
}
