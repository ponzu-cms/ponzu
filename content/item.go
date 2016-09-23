package content

// Item should only be embedded into content type structs.
type Item struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
}
