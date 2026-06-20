// Package gateway is the catalog outbound adapter for notifications.
package gateway

import (
	"log"

	"archview-example-hexagonal/internal/catalog/domain"
	"archview-example-hexagonal/internal/catalog/port"
)

type itemNotifier struct{}

// NewItemNotifier returns an ItemNotifier that logs.
func NewItemNotifier() port.ItemNotifier { return &itemNotifier{} }

func (n *itemNotifier) NotifyItemCreated(it domain.Item) {
	log.Printf("item created: %d %s", it.ID, it.Name)
}
