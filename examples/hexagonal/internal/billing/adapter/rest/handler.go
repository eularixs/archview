// Package rest is the billing inbound adapter (HTTP/REST driving side).
package rest

import (
	"encoding/json"
	"net/http"

	"archview-example-hexagonal/internal/billing/port"
)

// BillingHandler serves billing endpoints over the inbound port.
type BillingHandler struct {
	svc port.InvoiceService
}

// NewBillingHandler builds the adapter.
func NewBillingHandler(svc port.InvoiceService) *BillingHandler {
	return &BillingHandler{svc: svc}
}

// ListInvoices handles GET /billing/invoices.
func (h *BillingHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(h.svc.ListInvoices())
}

// CreateInvoice handles POST /billing/invoices.
func (h *BillingHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Customer string `json:"customer"`
		Amount   int    `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(h.svc.CreateInvoice(body.Customer, body.Amount))
}
