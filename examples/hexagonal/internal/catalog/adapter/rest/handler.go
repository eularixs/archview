// Package rest is the catalog inbound adapter (HTTP/REST driving side).
package rest

import (
	"encoding/json"
	"net/http"

	"archview-example-hexagonal/internal/catalog/port"
)

// CatalogHandler serves catalog endpoints over the inbound port.
type CatalogHandler struct {
	svc port.ItemService
}

// NewCatalogHandler builds the adapter.
func NewCatalogHandler(svc port.ItemService) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

// ListItems handles GET /catalog/items.
func (h *CatalogHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(h.svc.ListItems())
}

// CreateItem handles POST /catalog/items.
func (h *CatalogHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(h.svc.CreateItem(body.Name, body.Price))
}
