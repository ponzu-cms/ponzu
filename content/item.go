package content

import (
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

// Sluggable makes a struct locatable by URL with it's own path
// As an Item implementing Sluggable, slugs may overlap. If this is an issue,
// make your content struct (or one which imbeds Item) implement Sluggable
// and it will override the slug created by Item's SetSlug with your struct's
type Sluggable interface {
	SetSlug(string)
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

// Hookable provides our user with an easy way to intercept or add functionality
// to the different lifecycles/events a struct may encounter. Item implements
// Hookable with no-ops so our user can override only whichever ones necessary.
type Hookable interface {
	BeforeSave(req *http.Request) error
	AfterSave(req *http.Request) error

	BeforeDelete(req *http.Request) error
	AfterDelete(req *http.Request) error

	BeforeApprove(req *http.Request) error
	AfterApprove(req *http.Request) error

	BeforeReject(req *http.Request) error
	AfterReject(req *http.Request) error
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
func (i Item) BeforeSave(req *http.Request) error {
	return nil
}

// AfterSave is a no-op to ensure structs which embed Item implement Hookable
func (i Item) AfterSave(req *http.Request) error {
	return nil
}

// BeforeDelete is a no-op to ensure structs which embed Item implement Hookable
func (i Item) BeforeDelete(req *http.Request) error {
	return nil
}

// AfterDelete is a no-op to ensure structs which embed Item implement Hookable
func (i Item) AfterDelete(req *http.Request) error {
	return nil
}

// BeforeApprove is a no-op to ensure structs which embed Item implement Hookable
func (i Item) BeforeApprove(req *http.Request) error {
	return nil
}

// AfterApprove is a no-op to ensure structs which embed Item implement Hookable
func (i Item) AfterApprove(req *http.Request) error {
	return nil
}

// BeforeReject is a no-op to ensure structs which embed Item implement Hookable
func (i Item) BeforeReject(req *http.Request) error {
	return nil
}

// AfterReject is a no-op to ensure structs which embed Item implement Hookable
func (i Item) AfterReject(req *http.Request) error {
	return nil
}
