// Package domain holds the catalog entities (the hexagon core).
package domain

// Item is a catalog item.
type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}
