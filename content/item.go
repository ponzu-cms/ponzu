package content

// Item should only be embedded into content type structs.
// Helper for DB-related actions
type Item struct {
	ID int `json:"id"`
}
