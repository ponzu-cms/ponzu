package admin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func storeFileUploads(req *http.Request) (map[string]string, error) {
	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	ts := req.FormValue("timestamp")

	// To use for FormValue name:urlPath
	urlPaths := make(map[string]string)

	// get ts values individually to use as directory names when storing
	// uploaded images
	date := make(map[string]int)
	if ts == "" {
		now := time.Now()
		date["year"] = now.Year()
		date["month"] = int(now.Month())
		date["day"] = now.Day()

		// create timestamp format 'yyyy-mm-dd' and set in PostForm for
		// db insertion
		ts = fmt.Sprintf("%d-%02d-%02d", date["year"], date["month"], date["day"])
		req.PostForm.Set("timestamp", ts)
	} else {
		tsParts := strings.Split(ts, "-")
		year, err := strconv.Atoi(tsParts[0])
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}

		month, err := strconv.Atoi(tsParts[1])
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}

		day, err := strconv.Atoi(tsParts[2])
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}

		date["year"] = year
		date["month"] = month
		date["day"] = day
	}

	// get or create upload directory to save files from request
	pwd, err := os.Getwd()
	if err != nil {
		err := fmt.Errorf("Failed to locate current directory: %s", err)
		return nil, err
	}

	tsParts := strings.Split(ts, "-")
	urlPathPrefix := "api"
	uploadDirName := "uploads"

	uploadDir := filepath.Join(pwd, uploadDirName, tsParts[0], tsParts[1])
	err = os.MkdirAll(uploadDir, os.ModeDir|os.ModePerm)

	// loop over all files and save them to disk
	for name, fds := range req.MultipartForm.File {
		filename := fds[0].Filename
		src, err := fds[0].Open()
		if err != nil {
			err := fmt.Errorf("Couldn't open uploaded file: %s", err)
			return nil, err

		}
		defer src.Close()

		// check if file at path exists, if so, add timestamp to file
		absPath := filepath.Join(uploadDir, filename)

		if _, err := os.Stat(absPath); !os.IsNotExist(err) {
			filename = fmt.Sprintf("%d-%s", time.Now().Unix(), filename)
			absPath = filepath.Join(uploadDir, filename)
		}

		// save to disk (TODO: or check if S3 credentials exist, & save to cloud)
		dst, err := os.Create(absPath)
		if err != nil {
			err := fmt.Errorf("Failed to create destination file for upload: %s", err)
			return nil, err
		}

		// copy file from src to dst on disk
		if _, err = io.Copy(dst, src); err != nil {
			err := fmt.Errorf("Failed to copy uploaded file to destination: %s", err)
			return nil, err
		}

		// add name:urlPath to req.PostForm to be inserted into db
		urlPath := fmt.Sprintf("/%s/%s/%s/%s/%s", urlPathPrefix, uploadDirName, tsParts[0], tsParts[1], filename)

		urlPaths[name] = urlPath
	}

	return urlPaths, nil
}
