package root

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fanky5g/ponzu/internal/domain/entities/item"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	sortContentCalls = make(map[string]time.Time)
	waitDuration     = time.Millisecond * 2000
	sortMutex        = &sync.Mutex{}
)

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

func (repo *repository) enoughTime(key string) bool {
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
			repo.Sort(key)
		}
	}()

	return false
}

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

// Sort sorts all content of the type supplied as the namespace by time,
// in descending order, from most recent to least recent
// Should be called from a goroutine after SetContent is successful
func (repo *repository) Sort(namespace string) error {
	// wait if running too frequently per namespace
	if !repo.enoughTime(namespace) {
		return nil
	}

	// only sort main content types i.e. Post
	if strings.Contains(namespace, "__") {
		return nil
	}

	all, err := repo.FindAll(namespace)
	if err != nil {
		return err
	}

	var posts sortableContent
	// decode each (json) into type to then sort
	for i := range all {
		if sortable, ok := all[i].(item.Sortable); ok {
			posts = append(posts, sortable)
		}
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
			return err
		}

		bb = append(bb, j)
	}

	// store in <namespace>_sorted bucket, first delete existing
	err = repo.db.Update(func(tx *bolt.Tx) error {
		bname := []byte(namespace + "__sorted")
		err := tx.DeleteBucket(bname)
		if err != nil && !errors.Is(err, bolt.ErrBucketNotFound) {
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
		return err
	}

	return nil
}
