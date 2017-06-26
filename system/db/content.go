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
	"sync"
	"time"

	"github.com/ponzu-cms/ponzu/system/item"
	"github.com/ponzu-cms/ponzu/system/search"

	"github.com/boltdb/bolt"
	"github.com/gorilla/schema"
	uuid "github.com/satori/go.uuid"
)

// IsValidID checks that an ID from a DB target is valid.
// ID should be an integer greater than 0.
// ID of -1 is special for new posts, not updates.
// IDs start at 1 for auto-incrementing
func IsValidID(id string) bool {
	if i, err := strconv.Atoi(id); err != nil || i < 1 {
		return false
	}
	return true
}

// SetContent inserts/replaces values in the database.
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

	return update(ns, id, data, nil)
}

// UpdateContent updates/merges values in the database.
// The `target` argument is a string made up of namespace:id (string:int)
func UpdateContent(target string, data url.Values) (int, error) {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	if !IsValidID(id) {
		return 0, fmt.Errorf("Invalid ID in target for UpdateContent: %s", target)
	}

	// retrieve existing content from the database
	existingContent, err := Content(target)
	if err != nil {
		return 0, err
	}
	return update(ns, id, data, &existingContent)
}

// update can support merge or replace behavior depending on existingContent.
// if existingContent is non-nil, we merge field values. empty/missing fields are ignored.
// if existingContent is nil, we replace field values. empty/missing fields are reset.
func update(ns, id string, data url.Values, existingContent *[]byte) (int, error) {
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

	var j []byte
	if existingContent == nil {
		j, err = postToJSON(ns, data)
		if err != nil {
			return 0, err
		}
	} else {
		j, err = mergeData(ns, data, *existingContent)
		if err != nil {
			return 0, err
		}
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

	go func() {
		// update data in search index
		target := fmt.Sprintf("%s:%s", ns, id)
		err = search.UpdateIndex(target, j)
		if err != nil {
			log.Println("[search] UpdateIndex Error:", err)
		}
	}()

	return cid, nil
}

func mergeData(ns string, data url.Values, existingContent []byte) ([]byte, error) {
	var j []byte
	t, ok := item.Types[ns]
	if !ok {
		return nil, fmt.Errorf("Namespace type not found: %s", ns)
	}

	// Unmarsal the existing values
	s := t()
	err := json.Unmarshal(existingContent, &s)
	if err != nil {
		log.Println("Error decoding json while updating", ns, ":", err)
		return j, err
	}

	// Don't allow the Item fields to be updated from form values
	data.Del("id")
	data.Del("uuid")
	data.Del("slug")

	dec := schema.NewDecoder()
	dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
	dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
	err = dec.Decode(s, data)
	if err != nil {
		return j, err
	}

	j, err = json.Marshal(s)
	if err != nil {
		return j, err
	}

	return j, nil
}

func insert(ns string, data url.Values) (int, error) {
	var effectedID int
	var specifier string // i.e. __pending, __sorted, etc.
	if strings.Contains(ns, "__") {
		spec := strings.Split(ns, "__")
		ns = spec[0]
		specifier = "__" + spec[1]
	}

	var j []byte
	var cid string
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
		cid = strconv.FormatUint(id, 10)
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

		j, err = postToJSON(ns, data)
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

	go func() {
		// add data to search index
		target := fmt.Sprintf("%s:%s", ns, cid)
		err = search.UpdateIndex(target, j)
		if err != nil {
			log.Println("[search] UpdateIndex Error:", err)
		}
	}()

	return effectedID, nil
}

// DeleteContent removes an item from the database. Deleting a non-existent item
// will return a nil error.
func DeleteContent(target string) error {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	b, err := Content(target)
	if err != nil {
		return err
	}

	// get content slug to delete from __contentIndex if it exists
	// this way content added later can use slugs even if previously
	// deleted content had used one
	var itm item.Item
	err = json.Unmarshal(b, &itm)
	if err != nil {
		return err
	}

	err = store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		err := b.Delete([]byte(id))
		if err != nil {
			return err
		}

		// if content has a slug, also delete it from __contentIndex
		if itm.Slug != "" {
			ci := tx.Bucket([]byte("__contentIndex"))
			if ci == nil {
				return bolt.ErrBucketNotFound
			}

			err := ci.Delete([]byte(itm.Slug))
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

	go func() {
		// delete indexed data from search index
		if !strings.Contains(ns, "__") {
			target = fmt.Sprintf("%s:%s", ns, id)
			err = search.DeleteIndex(target)
			if err != nil {
				log.Println("[search] DeleteIndex Error:", err)
			}
		}
	}()

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

// ContentMulti returns a set of content based on the the targets / identifiers
// provided in Ponzu target string format: Type:ID
// NOTE: All targets should be of the same type
func ContentMulti(targets []string) ([][]byte, error) {
	var contents [][]byte
	for i := range targets {
		b, err := Content(targets[i])
		if err != nil {
			return nil, err
		}

		contents = append(contents, b)
	}

	return contents, nil
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
		case "desc", "":
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

		case "asc":
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
		}

		return nil
	})

	return total, posts
}

var sortContentCalls = make(map[string]time.Time)
var waitDuration = time.Millisecond * 2000
var sortMutex = &sync.Mutex{}

func setLastInvocation(key string) {
	sortMutex.Lock()
	sortContentCalls[key] = time.Now()
	sortMutex.Unlock()
}

func lastInvocation(key string) (time.Time, bool) {
	sortMutex.Lock()
	last, ok := sortContentCalls[key]
	sortMutex.Unlock()
	return last, ok
}

func enoughTime(key string) bool {
	last, ok := lastInvocation(key)
	if !ok {
		// no invocation yet
		// track next invocation
		setLastInvocation(key)
		return true
	}

	// if our required wait time has been met, return true
	if time.Now().After(last.Add(waitDuration)) {
		setLastInvocation(key)
		return true
	}

	// dispatch a delayed invocation in case no additional one follows
	go func() {
		lastInvocationBeforeTimer, _ := lastInvocation(key) // zero value can be handled, no need for ok
		enoughTimer := time.NewTimer(waitDuration)
		<-enoughTimer.C
		lastInvocationAfterTimer, _ := lastInvocation(key)
		if !lastInvocationAfterTimer.After(lastInvocationBeforeTimer) {
			SortContent(key)
		}
	}()

	return false
}

// SortContent sorts all content of the type supplied as the namespace by time,
// in descending order, from most recent to least recent
// Should be called from a goroutine after SetContent is successful
func SortContent(namespace string) {
	// wait if running too frequently per namespace
	if !enoughTime(namespace) {
		return
	}

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

		// encode to json and store as 'post.Time():i':post
		for i := range bb {
			cid := fmt.Sprintf("%d:%d", posts[i].Time(), i)
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

	// check for any multi-value fields (ex. checkbox fields)
	// and correctly format for db storage. Essentially, we need
	// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
	fieldOrderValue := make(map[string]map[string][]string)
	for k, v := range data {
		if strings.Contains(k, ".") {
			fo := strings.Split(k, ".")

			// put the order and the field value into map
			field := string(fo[0])
			order := string(fo[1])
			if len(fieldOrderValue[field]) == 0 {
				fieldOrderValue[field] = make(map[string][]string)
			}

			// orderValue is 0:[?type=Thing&id=1]
			orderValue := fieldOrderValue[field]
			orderValue[order] = v
			fieldOrderValue[field] = orderValue

			// discard the post form value with name.N
			data.Del(k)
		}
	}

	// add/set the key & value to the post form in order
	for f, ov := range fieldOrderValue {
		for i := 0; i < len(ov); i++ {
			position := fmt.Sprintf("%d", i)
			fieldValue := ov[position]

			if data.Get(f) == "" {
				for i, fv := range fieldValue {
					if i == 0 {
						data.Set(f, fv)
					} else {
						data.Add(f, fv)
					}
				}
			} else {
				for _, fv := range fieldValue {
					data.Add(f, fv)
				}
			}
		}
	}

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
