package config

import (
	"os"
	"path/filepath"
	"runtime"
)

var rootPath string

func init() {
	_, b, _, _ := runtime.Caller(0)
	rootPath = filepath.Join(filepath.Dir(b), "../..")
}

func DataDir() string {
	dataDir := os.Getenv("PONZU_DATA_DIR")
	if dataDir == "" {
		return rootPath
	}

	return dataDir
}

func TlsDir() string {
	tlsDir := os.Getenv("PONZU_TLS_DIR")
	if tlsDir == "" {
		tlsDir = filepath.Join(rootPath, "internal", "tls")
	}

	return tlsDir
}

func AdminStaticDir() string {
	staticDir := os.Getenv("PONZU_ADMINSTATIC_DIR")
	if staticDir == "" {
		staticDir = filepath.Join(rootPath, "public", "static")
	}

	return staticDir
}

func UploadDir() string {
	uploadDir := os.Getenv("PONZU_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = filepath.Join(DataDir(), "uploads")
	}
	return uploadDir
}

func SearchDir() string {
	searchDir := os.Getenv("PONZU_SEARCH_DIR")
	if searchDir == "" {
		searchDir = filepath.Join(DataDir(), "search")
	}
	return searchDir
}
