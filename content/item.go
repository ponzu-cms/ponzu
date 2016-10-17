package content

import "time"

// Item should only be embedded into content type structs.
type Item struct {
	ID        int       `json:"id"`
	Slug      string    `json:"slug"`
	Timestamp time.Time `json:"timestamp"`
	Updated   time.Time `json:"updated"`
}
