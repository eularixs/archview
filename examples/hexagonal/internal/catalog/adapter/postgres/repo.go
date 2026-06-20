// Package postgres is the catalog outbound adapter for persistence.
package postgres

import (
	"archview-example-hexagonal/internal/catalog/domain"
	"archview-example-hexagonal/internal/catalog/port"
)

type itemRepository struct {
	rows []domain.Item
	next int
}

// NewItemRepository returns an in-memory ItemRepository.
func NewItemRepository() port.ItemRepository {
	return &itemRepository{rows: []domain.Item{{ID: 1, Name: "Keyboard", Price: 40}}, next: 2}
}

func (r *itemRepository) FindAllItems() []domain.Item { return r.rows }

func (r *itemRepository) InsertItem(it domain.Item) domain.Item {
	it.ID = r.next
	r.next++
	r.rows = append(r.rows, it)
	return it
}
