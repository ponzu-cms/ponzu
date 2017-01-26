package item

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Sluggable makes a struct locatable by URL with it's own path.
// As an Item implementing Sluggable, slugs may overlap. If this is an issue,
// make your content struct (or one which embeds Item) implement Sluggable
// and it will override the slug created by Item's SetSlug with your own
type Sluggable interface {
	SetSlug(string)
	ItemSlug() string
}

// Identifiable enables a struct to have its ID set/get. Typically this is done
// to set an ID to -1 indicating it is new for DB inserts, since by default
// a newly initialized struct would have an ID of 0, the int zero-value, and
// BoltDB's starting key per bucket is 0, thus overwriting the first record.
type Identifiable interface {
	ItemID() int
	SetItemID(int)
	UniqueID() uuid.UUID
	String() string
}

// Sortable ensures data is sortable by time
type Sortable interface {
	Time() int64
	Touch() int64
}

// Hookable provides our user with an easy way to intercept or add functionality
// to the different lifecycles/events a struct may encounter. Item implements
// Hookable with no-ops so our user can override only whichever ones necessary.
type Hookable interface {
	BeforeSave(http.ResponseWriter, *http.Request) error
	AfterSave(http.ResponseWriter, *http.Request) error

	BeforeDelete(http.ResponseWriter, *http.Request) error
	AfterDelete(http.ResponseWriter, *http.Request) error

	BeforeApprove(http.ResponseWriter, *http.Request) error
	AfterApprove(http.ResponseWriter, *http.Request) error

	BeforeReject(http.ResponseWriter, *http.Request) error
	AfterReject(http.ResponseWriter, *http.Request) error
}

// Hideable lets a user keep items hidden
type Hideable interface {
	Hide(http.ResponseWriter, *http.Request) error
}

// Pushable lets a user define which values of certain struct fields are
// 'pushed' down to  a client via HTTP/2 Server Push. All items in the slice
// should be the json tag names of the struct fields to which they coorespond.
type Pushable interface {
	// the values contained by fields returned by Push must strictly be URL paths
	Push() []string
}

// Item should only be embedded into content type structs.
type Item struct {
	UUID      uuid.UUID `json:"uuid"`
	ID        int       `json:"id"`
	Slug      string    `json:"slug"`
	Timestamp int64     `json:"timestamp"`
	Updated   int64     `json:"updated"`
}

// Time partially implements the Sortable interface
func (i Item) Time() int64 {
	return i.Timestamp
}

// Touch partially implements the Sortable interface
func (i Item) Touch() int64 {
	return i.Updated
}

// SetSlug sets the item's slug for its URL
func (i *Item) SetSlug(slug string) {
	i.Slug = slug
}

// ItemSlug sets the item's slug for its URL
func (i *Item) ItemSlug() string {
	return i.Slug
}

// ItemID gets the Item's ID field
// partially implements the Identifiable interface
func (i Item) ItemID() int {
	return i.ID
}

// SetItemID sets the Item's ID field
// partially implements the Identifiable interface
func (i *Item) SetItemID(id int) {
	i.ID = id
}

// UniqueID gets the Item's UUID field
// partially implements the Identifiable interface
func (i Item) UniqueID() uuid.UUID {
	return i.UUID
}

// String formats an Item into a printable value
// partially implements the Identifiable interface
func (i Item) String() string {
	return fmt.Sprintf("Item ID: %s", i.UniqueID())
}

// BeforeSave is a no-op to ensure structs which embed Item implement Hookable
func (i Item) BeforeSave(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterSave is a no-op to ensure structs which embed Item implement Hookable
func (i Item) AfterSave(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeDelete is a no-op to ensure structs which embed Item implement Hookable
func (i Item) BeforeDelete(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterDelete is a no-op to ensure structs which embed Item implement Hookable
func (i Item) AfterDelete(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeApprove is a no-op to ensure structs which embed Item implement Hookable
func (i Item) BeforeApprove(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterApprove is a no-op to ensure structs which embed Item implement Hookable
func (i Item) AfterApprove(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeReject is a no-op to ensure structs which embed Item implement Hookable
func (i Item) BeforeReject(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterReject is a no-op to ensure structs which embed Item implement Hookable
func (i Item) AfterReject(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// Slug returns a URL friendly string from the title of a post item
func Slug(i Identifiable) (string, error) {
	// get the name of the post item
	name := strings.TrimSpace(i.String())

	// filter out non-alphanumeric character or non-whitespace
	slug, err := stringToSlug(name)
	if err != nil {
		return "", err
	}

	return slug, nil
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

// modified version of: https://www.socketloop.com/tutorials/golang-format-strings-to-seo-friendly-url-example
func stringToSlug(s string) (string, error) {
	src := []byte(strings.ToLower(s))

	// convert all spaces to dash
	rx := regexp.MustCompile("[[:space:]]")
	src = rx.ReplaceAll(src, []byte("-"))

	// remove all blanks such as tab
	rx = regexp.MustCompile("[[:blank:]]")
	src = rx.ReplaceAll(src, []byte(""))

	rx = regexp.MustCompile("[!/:-@[-`{-~]")
	src = rx.ReplaceAll(src, []byte(""))

	rx = regexp.MustCompile("/[^\x20-\x7F]/")
	src = rx.ReplaceAll(src, []byte(""))

	rx = regexp.MustCompile("`&(amp;)?#?[a-z0-9]+;`i")
	src = rx.ReplaceAll(src, []byte("-"))

	rx = regexp.MustCompile("`&([a-z])(acute|uml|circ|grave|ring|cedil|slash|tilde|caron|lig|quot|rsquo);`i")
	src = rx.ReplaceAll(src, []byte("\\1"))

	rx = regexp.MustCompile("`[^a-z0-9]`i")
	src = rx.ReplaceAll(src, []byte("-"))

	rx = regexp.MustCompile("`[-]+`")
	src = rx.ReplaceAll(src, []byte("-"))

	str := strings.Replace(string(src), "'", "", -1)
	str = strings.Replace(str, `"`, "", -1)

	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	slug, _, err := transform.String(t, str)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(slug), nil
}
