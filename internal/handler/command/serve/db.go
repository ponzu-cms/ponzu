package serve

import (
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/config"
	"log"
	"path/filepath"
)

var store *bolt.DB

func GetDatabase() *bolt.DB {
	if store == nil {
		var err error
		systemDb := filepath.Join(config.DataDir(), "system.db")
		store, err = bolt.Open(systemDb, 0666, nil)
		if err != nil {
			log.Fatalln(err)
		}
	}

	return store
}

func Close(store *bolt.DB) {
	err := store.Close()
	if err != nil {
		log.Println(err)
	}
}
