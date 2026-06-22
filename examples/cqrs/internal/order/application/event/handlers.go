// Package event holds order event subscribers (CQRS read-model / side effects).
package event

import (
	"context"
	"log"

	"archview-example-cqrs/internal/platform/cqrs"
)

// EmailOnOrderCreated sends a confirmation email.
type EmailOnOrderCreated struct{}

// Handle satisfies cqrs.EventHandler.
func (EmailOnOrderCreated) Handle(ctx context.Context, e cqrs.Event) error {
	log.Printf("email: %s", e.EventName())
	return nil
}

// AdjustInventory updates stock levels.
type AdjustInventory struct{}

// Handle satisfies cqrs.EventHandler.
func (AdjustInventory) Handle(ctx context.Context, e cqrs.Event) error {
	log.Printf("inventory: %s", e.EventName())
	return nil
}
