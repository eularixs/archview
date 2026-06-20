// Package port declares the billing hexagon's inbound and outbound boundaries.
package port

import "archview-example-hexagonal/internal/billing/domain"

// InvoiceService is the inbound port. Impl: usecase.invoiceService.
type InvoiceService interface {
	ListInvoices() []domain.Invoice
	CreateInvoice(customer string, amount int) domain.Invoice
}

// InvoiceRepository is the outbound persistence port. Impl: postgres.invoiceRepository.
type InvoiceRepository interface {
	FindAllInvoices() []domain.Invoice
	InsertInvoice(in domain.Invoice) domain.Invoice
}
