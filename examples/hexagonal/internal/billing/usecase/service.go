// Package usecase implements the billing inbound port over its outbound port.
package usecase

import (
	"archview-example-hexagonal/internal/billing/domain"
	"archview-example-hexagonal/internal/billing/port"
)

type invoiceService struct {
	repo port.InvoiceRepository
}

// NewInvoiceService wires the use case over its repository port.
func NewInvoiceService(repo port.InvoiceRepository) port.InvoiceService {
	return &invoiceService{repo: repo}
}

func (s *invoiceService) ListInvoices() []domain.Invoice {
	return s.repo.FindAllInvoices()
}

func (s *invoiceService) CreateInvoice(customer string, amount int) domain.Invoice {
	return s.repo.InsertInvoice(domain.Invoice{Customer: customer, Amount: amount})
}
