package content

// Item should only be embedded into content type structs.
type Item struct {
	ID        int    `json:"id"`
	Slug      string `json:"slug"`
	Timestamp int64  `json:"timestamp"`
	Updated   int64  `json:"updated"`
}

// Time partially implements the Sortable interface
func (i Item) Time() int64 {
	return i.Timestamp
}

// Touch partially implements the Sortable interface
func (i Item) Touch() int64 {
	return i.Updated
}

// ItemID partially implements the Sortable interface
func (i Item) ItemID() int {
	return i.ID
}

// SetSlug sets the item's slug for its URL
func (i *Item) SetSlug(slug string) {
	i.Slug = slug
}

// Sluggable makes a struct locatable by URL with it's own path
// As an Item implementing Sluggable, slugs may overlap. If this is an issue,
// make your content struct (or one which imbeds Item) implement Sluggable
// and it will override the slug created by Item's SetSlug with your struct's
type Sluggable interface {
	SetSlug(string)
}

// Identifiable enables a struct to have its ID set. Typically this is done
// to set an ID to -1 indicating it is new for DB inserts, since by default
// a newly initialized struct would have an ID of 0, the int zero-value, and
// BoltDB's starting key per bucket is 0, thus overwriting the first record.
type Identifiable interface {
	SetContentID(int)
}
