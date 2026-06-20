// Package domain holds the billing entities (the hexagon core).
package domain

// Invoice is a billing invoice.
type Invoice struct {
	ID       int    `json:"id"`
	Customer string `json:"customer"`
	Amount   int    `json:"amount"`
}
