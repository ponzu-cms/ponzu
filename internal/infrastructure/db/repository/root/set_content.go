package root

import (
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"log"
	"strconv"
	"strings"
)

// SetEntity inserts/replaces values in the database.
// The `target` argument is a string made up of namespace:id (string:int)
func (repo *repository) SetEntity(ns string, data interface{}) (string, error) {
	var specifier string // i.e. __pending, __sorted, etc.
	if strings.Contains(ns, "__") {
		spec := strings.Split(ns, "__")
		ns = spec[0]
		specifier = "__" + spec[1]
	}

	identifiable, ok := data.(item.Identifiable)
	if !ok {
		return "", errors.New("item does not implement identifiable interface")
	}

	cid := identifiable.ItemID()
	err := repo.db.Update(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		b, err := tx.CreateBucketIfNotExists([]byte(ns + specifier))
		if err != nil {
			return err
		}

		if cid == "" {
			id, err := b.NextSequence()
			if err != nil {
				return err
			}

			cid = strconv.FormatUint(id, 10)
			data.(item.Identifiable).SetItemID(cid)
		}

		j, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return b.Put([]byte(cid), j)
	})

	if err != nil {
		log.Println(err)
		return "", err
	}

	// store the slug,type:id in contentIndex if public content
	// TODO: if isNew store in content index
	//if specifier == "" {
	//	ci := tx.Bucket([]byte("__contentIndex"))
	//	if ci == nil {
	//		return bolt.ErrBucketNotFound
	//	}
	//
	//	k := []byte(data.Get("slug"))
	//	v := []byte(fmt.Sprintf("%s:%d", ns, effectedID))
	//	err := ci.Put(k, v)
	//	if err != nil {
	//		return err
	//	}
	//}

	if specifier == "" {
		go func() {
			err = repo.Sort(ns)
			if err != nil {
				log.Println(err)
			}
		}()
	}

	return cid, nil
}
