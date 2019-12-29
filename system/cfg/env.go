package cfg

import (
	"log"
	"os"
	"path/filepath"
)

func getWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't find working directory", err)
	}
	return wd
}

func DataDir() string {
	dataDir := os.Getenv("PONZU_DATA_DIR")
	if dataDir == "" {
		return getWd()
	}
	return dataDir
}

func TlsDir() string {
	tlsDir := os.Getenv("PONZU_TLS_DIR")
	if tlsDir == "" {
		tlsDir = filepath.Join(getWd(), "cmd", "ponzu", "vendor", "github.com", "ponzu-cms", "ponzu", "system", "tls")
	}
	return tlsDir
}

func AdminStaticDir() string {
	staticDir := os.Getenv("PONZU_ADMINSTATIC_DIR")
	if staticDir == "" {

		staticDir = filepath.Join(getWd(), "cmd", "ponzu", "vendor", "github.com", "ponzu-cms", "ponzu", "system", "admin", "static")
	}
	return staticDir
}

func UploadDir() string {
	uploadDir := os.Getenv("PONZU_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = filepath.Join(DataDir(),"uploads")
	}
	return uploadDir
}

func SearchDir() string {
	searchDir := os.Getenv("PONZU_SEARCH_DIR")
	if searchDir == "" {
		searchDir = filepath.Join(DataDir(),"search")
	}
	return searchDir
}
