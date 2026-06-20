// Package postgres is the billing outbound adapter for persistence.
package postgres

import (
	"archview-example-hexagonal/internal/billing/domain"
	"archview-example-hexagonal/internal/billing/port"
)

type invoiceRepository struct {
	rows []domain.Invoice
	next int
}

// NewInvoiceRepository returns an in-memory InvoiceRepository.
func NewInvoiceRepository() port.InvoiceRepository {
	return &invoiceRepository{rows: []domain.Invoice{{ID: 1, Customer: "Acme", Amount: 100}}, next: 2}
}

func (r *invoiceRepository) FindAllInvoices() []domain.Invoice { return r.rows }

func (r *invoiceRepository) InsertInvoice(in domain.Invoice) domain.Invoice {
	in.ID = r.next
	r.next++
	r.rows = append(r.rows, in)
	return in
}
