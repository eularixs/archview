// Package usecase implements the catalog inbound port over its outbound ports.
package usecase

import (
	"archview-example-hexagonal/internal/catalog/domain"
	"archview-example-hexagonal/internal/catalog/port"
)

type itemService struct {
	repo     port.ItemRepository
	notifier port.ItemNotifier
}

// NewItemService wires the use case over its outbound ports.
func NewItemService(repo port.ItemRepository, notifier port.ItemNotifier) port.ItemService {
	return &itemService{repo: repo, notifier: notifier}
}

func (s *itemService) ListItems() []domain.Item {
	return s.repo.FindAllItems()
}

func (s *itemService) CreateItem(name string, price int) domain.Item {
	it := s.repo.InsertItem(domain.Item{Name: name, Price: price})
	s.notifier.NotifyItemCreated(it)
	return it
}
