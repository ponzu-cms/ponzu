// Package db ...
package db

// Item should only be embedded into content type structs.
// Helper for DB-related actions
type Item struct {
	ID int `json:"_id"`
}
