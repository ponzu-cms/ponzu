package analytics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
)

// Backup writes a snapshot of the system.db database to an HTTP response. The
// output is discarded if we get a cancellation signal.
func Backup(ctx context.Context, res http.ResponseWriter) error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- store.View(func(tx *bolt.Tx) error {
			ts := time.Now().Unix()
			disposition := `attachment; filename="analytics-%d.db.bak"`

			res.Header().Set("Content-Type", "application/octet-stream")
			res.Header().Set("Content-Disposition", fmt.Sprintf(disposition, ts))
			res.Header().Set("Content-Length", fmt.Sprintf("%d", int(tx.Size())))

			_, err := tx.WriteTo(res)
			return err
		})
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
