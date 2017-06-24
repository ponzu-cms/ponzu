package search

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ponzu-cms/ponzu/system/backup"
)

// Backup creates an archive of a project's search index and writes it
// to the response as a download
func Backup(ctx context.Context, res http.ResponseWriter) error {
	ts := time.Now().Unix()
	filename := fmt.Sprintf("search-%d.bak.tar.gz", ts)
	tmp := os.TempDir()
	bk := filepath.Join(tmp, filename)

	// create search-{stamp}.bak.tar.gz
	f, err := os.Create(bk)
	if err != nil {
		return err
	}

	backup.ArchiveFS(ctx, "search", f)

	err = f.Close()
	if err != nil {
		return err
	}

	// write data to response
	data, err := os.Open(bk)
	if err != nil {
		return err
	}
	defer data.Close()
	defer os.Remove(bk)

	disposition := `attachment; filename=%s`
	info, err := data.Stat()
	if err != nil {
		return err
	}

	res.Header().Set("Content-Type", "application/octet-stream")
	res.Header().Set("Content-Disposition", fmt.Sprintf(disposition, ts))
	res.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	_, err = io.Copy(res, data)

	return err
}
