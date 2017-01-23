package upload

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Backup creates an archive of a project's uploads and writes it
// to the response as a download
func Backup(res http.ResponseWriter) error {
	ts := time.Now().Unix()
	filename := fmt.Sprintf("uploads-%d.bak.tar.gz", ts)
	tmp := os.TempDir()

	// create uploads-{stamp}.bak.tar.gz
	f, err := os.Create(filepath.Join(tmp, filename))
	if err != nil {
		return err
	}
	defer f.Close()

	// loop through directory and gzip files
	// add all to uploads.bak.tar.gz tarball
	gz := gzip.NewWriter(f)
	tarball := tar.NewWriter(gz)
	err = filepath.Walk("uploads", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		h := &tar.Header{
			Name:    info.Name(),
			Size:    info.Size(),
			Mode:    int64(info.Mode()),
			ModTime: info.ModTime(),
		}

		err = tarball.WriteHeader(h)
		if err != nil {
			return err
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(tarball, src)

		return err
	})

	// write data to response
	data, err := os.Open(filepath.Join(tmp, filename))
	if err != nil {
		return err
	}
	defer data.Close()

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
