package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ArchiveFS walks the filesystem starting from basedir writing files encountered
// tarred and gzipped to the provided writer
func ArchiveFS(ctx context.Context, basedir string, w io.Writer) error {
	gz := gzip.NewWriter(w)
	tarball := tar.NewWriter(gz)

	errChan := make(chan error, 1)
	walkFn := func(path string, info os.FileInfo, err error) error {
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
	}

	// stop processing if we get a cancellation signal
	err := filepath.Walk(basedir, func(path string, info os.FileInfo, err error) error {
		go func() { errChan <- walkFn(path, info, err) }()

		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return err
			}
		case err := <-errChan:
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

	return nil
}
