package root

import (
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"strings"
)

// DeleteEntity removes an item from the database. Deleting a non-existent item
// will return a nil error.
func (repo *repository) DeleteEntity(target string) error {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	itm, err := repo.FindOneByTarget(target)
	if err != nil {
		return err
	}

	// get content slug to delete from __contentIndex if it exists
	// this way content added later can use slugs even if previously
	// deleted content had used one

	err = repo.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		err := b.Delete([]byte(id))
		if err != nil {
			return err
		}

		// if content has a slug, also delete it from __contentIndex
		if sluggable, ok := itm.(item.Sluggable); ok {
			if sluggable.ItemSlug() != "" {
				ci := tx.Bucket([]byte(contentIndexName))
				if ci == nil {
					return bolt.ErrBucketNotFound
				}

				err := ci.Delete([]byte(sluggable.ItemSlug()))
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// exception to typical "run in goroutine" pattern:
	// we want to have an updated controllers view as soon as this is deleted, so
	// in some cases, the delete and redirect is faster than the sort,
	// thus still showing a deleted post in the controllers view.
	return repo.Sort(ns)
}
