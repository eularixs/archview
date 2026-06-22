// Package cqrs is the mediator: command, query and event buses that dispatch to
// handlers registered in maps at runtime. This runtime indirection is exactly
// what a static call graph cannot follow precisely.
package cqrs

import (
	"context"
	"fmt"
)

// --- commands ---

// Command is a write intent.
type Command interface{ CommandName() string }

// CommandHandler handles one command type.
type CommandHandler interface {
	Handle(ctx context.Context, c Command) error
}

// CommandBus routes a command to its handler by name.
type CommandBus struct {
	handlers map[string]CommandHandler
}

// NewCommandBus builds an empty command bus.
func NewCommandBus() *CommandBus { return &CommandBus{handlers: map[string]CommandHandler{}} }

// Register binds a command name to its handler (runtime wiring).
func (b *CommandBus) Register(name string, h CommandHandler) { b.handlers[name] = h }

// Dispatch looks up the handler by command name and invokes it.
func (b *CommandBus) Dispatch(ctx context.Context, c Command) error {
	h, ok := b.handlers[c.CommandName()]
	if !ok {
		return fmt.Errorf("cqrs: no handler for command %q", c.CommandName())
	}
	return h.Handle(ctx, c)
}

// --- queries ---

// Query is a read intent.
type Query interface{ QueryName() string }

// QueryHandler handles one query type.
type QueryHandler interface {
	Handle(ctx context.Context, q Query) (any, error)
}

// QueryBus routes a query to its handler by name.
type QueryBus struct {
	handlers map[string]QueryHandler
}

// NewQueryBus builds an empty query bus.
func NewQueryBus() *QueryBus { return &QueryBus{handlers: map[string]QueryHandler{}} }

// Register binds a query name to its handler.
func (b *QueryBus) Register(name string, h QueryHandler) { b.handlers[name] = h }

// Dispatch looks up the handler by query name and invokes it.
func (b *QueryBus) Dispatch(ctx context.Context, q Query) (any, error) {
	h, ok := b.handlers[q.QueryName()]
	if !ok {
		return nil, fmt.Errorf("cqrs: no handler for query %q", q.QueryName())
	}
	return h.Handle(ctx, q)
}

// --- events ---

// Event is a fact that happened.
type Event interface{ EventName() string }

// EventHandler reacts to an event.
type EventHandler interface {
	Handle(ctx context.Context, e Event) error
}

// EventBus fans an event out to its subscribers.
type EventBus struct {
	subs map[string][]EventHandler
}

// NewEventBus builds an empty event bus.
func NewEventBus() *EventBus { return &EventBus{subs: map[string][]EventHandler{}} }

// Subscribe adds a subscriber for an event name.
func (b *EventBus) Subscribe(name string, h EventHandler) { b.subs[name] = append(b.subs[name], h) }

// Publish invokes every subscriber of the event (runtime fan-out).
func (b *EventBus) Publish(ctx context.Context, e Event) {
	for _, h := range b.subs[e.EventName()] {
		_ = h.Handle(ctx, e)
	}
}
