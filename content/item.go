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

// ContentID partially implements the Sortable interface
func (i Item) ContentID() int {
	return i.ID
}
