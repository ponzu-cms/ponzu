package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/ponzu-cms/ponzu/system/item"

	"github.com/boltdb/bolt"
	"github.com/gorilla/schema"
	uuid "github.com/satori/go.uuid"
)

// SetContent inserts or updates values in the database.
// The `target` argument is a string made up of namespace:id (string:int)
func SetContent(target string, data url.Values) (int, error) {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	// check if content id == -1 (indicating new post).
	// if so, run an insert which will assign the next auto incremented int.
	// this is done because boltdb begins its bucket auto increment value at 0,
	// which is the zero-value of an int in the Item struct field for ID.
	// this is a problem when the original first post (with auto ID = 0) gets
	// overwritten by any new post, originally having no ID, defauting to 0.
	if id == "-1" {
		return insert(ns, data)
	}

	return update(ns, id, data)
}

func update(ns, id string, data url.Values) (int, error) {
	var specifier string // i.e. __pending, __sorted, etc.
	if strings.Contains(ns, "__") {
		spec := strings.Split(ns, "__")
		ns = spec[0]
		specifier = "__" + spec[1]
	}

	cid, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}

	j, err := postToJSON(ns, data)
	if err != nil {
		return 0, err
	}

	err = store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(ns + specifier))
		if err != nil {
			return err
		}

		err = b.Put([]byte(fmt.Sprintf("%d", cid)), j)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, nil
	}

	if specifier == "" {
		go SortContent(ns)
	}

	// update changes data, so invalidate client caching
	err = InvalidateCache()
	if err != nil {
		return 0, err
	}

	return cid, nil
}

func insert(ns string, data url.Values) (int, error) {
	var effectedID int
	var specifier string // i.e. __pending, __sorted, etc.
	if strings.Contains(ns, "__") {
		spec := strings.Split(ns, "__")
		ns = spec[0]
		specifier = "__" + spec[1]
	}

	err := store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(ns + specifier))
		if err != nil {
			return err
		}

		// get the next available ID and convert to string
		// also set effectedID to int of ID
		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		cid := strconv.FormatUint(id, 10)
		effectedID, err = strconv.Atoi(cid)
		if err != nil {
			return err
		}
		data.Set("id", cid)

		// add UUID to data for use in embedded Item
		uid := uuid.NewV4()
		data.Set("uuid", uid.String())

		// if type has a specifier, add it to data for downstream processing
		if specifier != "" {
			data.Set("__specifier", specifier)
		}

		j, err := postToJSON(ns, data)
		if err != nil {
			return err
		}

		err = b.Put([]byte(cid), j)
		if err != nil {
			return err
		}

		// store the slug,type:id in contentIndex if public content
		if specifier == "" {
			ci := tx.Bucket([]byte("__contentIndex"))
			if ci == nil {
				return bolt.ErrBucketNotFound
			}

			k := []byte(data.Get("slug"))
			v := []byte(fmt.Sprintf("%s:%d", ns, effectedID))
			err := ci.Put(k, v)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	if specifier == "" {
		go SortContent(ns)
	}

	// insert changes data, so invalidate client caching
	err = InvalidateCache()
	if err != nil {
		return 0, err
	}

	return effectedID, nil
}

// DeleteContent removes an item from the database. Deleting a non-existent item
// will return a nil error.
func DeleteContent(target string, data url.Values) error {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	err := store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		err := b.Delete([]byte(id))
		if err != nil {
			return err
		}

		// if content has a slug, also delete it from __contentIndex
		slug := data.Get("slug")
		if slug != "" {
			ci := tx.Bucket([]byte("__contentIndex"))
			if ci == nil {
				return bolt.ErrBucketNotFound
			}

			err := ci.Delete([]byte(slug))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// delete changes data, so invalidate client caching
	err = InvalidateCache()
	if err != nil {
		return err
	}

	// exception to typical "run in goroutine" pattern:
	// we want to have an updated admin view as soon as this is deleted, so
	// in some cases, the delete and redirect is faster than the sort,
	// thus still showing a deleted post in the admin view.
	SortContent(ns)

	return nil
}

// Content retrives one item from the database. Non-existent values will return an empty []byte
// The `target` argument is a string made up of namespace:id (string:int)
func Content(target string) ([]byte, error) {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	val := &bytes.Buffer{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		_, err := val.Write(b.Get([]byte(id)))
		if err != nil {
			log.Println(err)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return val.Bytes(), nil
}

// ContentBySlug does a lookup in the content index to find the type and id of
// the requested content. Subsequently, issues the lookup in the type bucket and
// returns the the type and data at that ID or nil if nothing exists.
func ContentBySlug(slug string) (string, []byte, error) {
	val := &bytes.Buffer{}
	var t, id string
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__contentIndex"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		idx := b.Get([]byte(slug))

		if idx != nil {
			tid := strings.Split(string(idx), ":")

			if len(tid) < 2 {
				return fmt.Errorf("Bad data in content index for slug: %s", slug)
			}

			t, id = tid[0], tid[1]
		}

		c := tx.Bucket([]byte(t))
		if c == nil {
			return bolt.ErrBucketNotFound
		}
		_, err := val.Write(c.Get([]byte(id)))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return t, nil, err
	}

	return t, val.Bytes(), nil
}

// ContentAll retrives all items from the database within the provided namespace
func ContentAll(namespace string) [][]byte {
	var posts [][]byte
	store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(namespace))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		numKeys := b.Stats().KeyN
		posts = make([][]byte, 0, numKeys)

		b.ForEach(func(k, v []byte) error {
			posts = append(posts, v)

			return nil
		})

		return nil
	})

	return posts
}

// QueryOptions holds options for a query
type QueryOptions struct {
	Count  int
	Offset int
	Order  string
}

// Query retrieves a set of content from the db based on options
// and returns the total number of content in the namespace and the content
func Query(namespace string, opts QueryOptions) (int, [][]byte) {
	var posts [][]byte
	var total int

	// correct bad input rather than return nil or error
	// similar to default case for opts.Order switch below
	if opts.Count < 0 {
		opts.Count = -1
	}

	if opts.Offset < 0 {
		opts.Offset = 0
	}

	store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(namespace))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		c := b.Cursor()
		n := b.Stats().KeyN
		total = n

		// return nil if no content
		if n == 0 {
			return nil
		}

		var start, end int
		switch opts.Count {
		case -1:
			start = 0
			end = n

		default:
			start = opts.Count * opts.Offset
			end = start + opts.Count
		}

		// bounds check on posts given the start & end count
		if start > n {
			start = n - opts.Count
		}
		if end > n {
			end = n
		}

		i := 0   // count of num posts added
		cur := 0 // count of num cursor moves
		switch opts.Order {
		case "asc":
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				if cur < start {
					cur++
					continue
				}

				if cur >= end {
					break
				}

				posts = append(posts, v)
				i++
				cur++
			}

		case "desc", "":
			for k, v := c.First(); k != nil; k, v = c.Next() {
				if cur < start {
					cur++
					continue
				}

				if cur >= end {
					break
				}

				posts = append(posts, v)
				i++
				cur++
			}

		default:
			// results for DESC order
			for k, v := c.First(); k != nil; k, v = c.Next() {
				if cur < start {
					cur++
					continue
				}

				if cur >= end {
					break
				}

				posts = append(posts, v)
				i++
				cur++
			}
		}

		return nil
	})

	return total, posts
}

// SortContent sorts all content of the type supplied as the namespace by time,
// in descending order, from most recent to least recent
// Should be called from a goroutine after SetContent is successful
func SortContent(namespace string) {
	// only sort main content types i.e. Post
	if strings.Contains(namespace, "__") {
		return
	}

	all := ContentAll(namespace)

	var posts sortableContent
	// decode each (json) into type to then sort
	for i := range all {
		j := all[i]
		post := item.Types[namespace]()

		err := json.Unmarshal(j, &post)
		if err != nil {
			log.Println("Error decoding json while sorting", namespace, ":", err)
			return
		}

		posts = append(posts, post.(item.Sortable))
	}

	// sort posts
	sort.Sort(posts)

	// marshal posts to json
	var bb [][]byte
	for i := range posts {
		j, err := json.Marshal(posts[i])
		if err != nil {
			// log error and kill sort so __sorted is not in invalid state
			log.Println("Error marshal post to json in SortContent:", err)
			return
		}

		bb = append(bb, j)
	}

	// store in <namespace>_sorted bucket, first delete existing
	err := store.Update(func(tx *bolt.Tx) error {
		bname := []byte(namespace + "__sorted")
		err := tx.DeleteBucket(bname)
		if err != nil && err != bolt.ErrBucketNotFound {
			return err
		}

		b, err := tx.CreateBucketIfNotExists(bname)
		if err != nil {
			return err
		}

		// encode to json and store as 'i:post.Time()':post
		for i := range bb {
			cid := fmt.Sprintf("%d:%d", i, posts[i].Time())
			err = b.Put([]byte(cid), bb[i])
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Println("Error while updating db with sorted", namespace, err)
	}

}

type sortableContent []item.Sortable

func (s sortableContent) Len() int {
	return len(s)
}

func (s sortableContent) Less(i, j int) bool {
	return s[i].Time() > s[j].Time()
}

func (s sortableContent) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func postToJSON(ns string, data url.Values) ([]byte, error) {
	// find the content type and decode values into it
	t, ok := item.Types[ns]
	if !ok {
		return nil, fmt.Errorf(item.ErrTypeNotRegistered.Error(), ns)
	}
	post := t()

	dec := schema.NewDecoder()
	dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
	dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
	err := dec.Decode(post, data)
	if err != nil {
		return nil, err
	}

	// if the content has no slug, and has no specifier, create a slug, check it
	// for duplicates, and add it to our values
	if data.Get("slug") == "" && data.Get("__specifier") == "" {
		slug, err := item.Slug(post.(item.Identifiable))
		if err != nil {
			return nil, err
		}

		slug, err = checkSlugForDuplicate(slug)
		if err != nil {
			return nil, err
		}

		post.(item.Sluggable).SetSlug(slug)
		data.Set("slug", slug)
	}

	// marshall content struct to json for db storage
	j, err := json.Marshal(post)
	if err != nil {
		return nil, err
	}

	return j, nil
}

func checkSlugForDuplicate(slug string) (string, error) {
	// check for existing slug in __contentIndex
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__contentIndex"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		original := slug
		exists := true
		i := 0
		for exists {
			s := b.Get([]byte(slug))
			if s == nil {
				exists = false
				return nil
			}

			i++
			slug = fmt.Sprintf("%s-%d", original, i)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return slug, nil
}
