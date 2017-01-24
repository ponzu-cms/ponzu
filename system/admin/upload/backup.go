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
	backup := filepath.Join(tmp, filename)

	// create uploads-{stamp}.bak.tar.gz
	f, err := os.Create(backup)
	if err != nil {
		return err
	}

	// loop through directory and gzip files
	// add all to uploads.bak.tar.gz tarball
	gz := gzip.NewWriter(f)
	tarball := tar.NewWriter(gz)

	err = filepath.Walk("uploads", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		hdr.Name = path

		err = tarball.WriteHeader(hdr)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			src, err := os.Open(path)
			if err != nil {
				return err
			}
			defer src.Close()
			_, err = io.Copy(tarball, src)
			if err != nil {
				return err
			}

			err = tarball.Flush()
			if err != nil {
				return err
			}

			err = gz.Flush()
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = gz.Close()
	if err != nil {
		return err
	}
	err = tarball.Close()
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	// write data to response
	data, err := os.Open(backup)
	if err != nil {
		return err
	}
	defer data.Close()
	defer os.Remove(backup)

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
