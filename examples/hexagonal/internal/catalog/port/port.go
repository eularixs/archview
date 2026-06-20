// Package port declares the catalog hexagon's boundaries: the inbound
// (driving) port the world calls, and the outbound (driven) ports it needs.
package port

import "archview-example-hexagonal/internal/catalog/domain"

// ItemService is the inbound port (use-case boundary). Impl: usecase.itemService.
type ItemService interface {
	ListItems() []domain.Item
	CreateItem(name string, price int) domain.Item
}

// ItemRepository is an outbound port for persistence. Impl: postgres.itemRepository.
type ItemRepository interface {
	FindAllItems() []domain.Item
	InsertItem(it domain.Item) domain.Item
}

// ItemNotifier is an outbound port for notifications. Impl: gateway.itemNotifier.
type ItemNotifier interface {
	NotifyItemCreated(it domain.Item)
}
