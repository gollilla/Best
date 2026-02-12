package events

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Emitter is a channel-based event emitter that replaces TypeScript's EventEmitter
type Emitter struct {
	listeners map[EventName]map[string]*listener
	mu        sync.RWMutex
}

// listener represents a single event listener
type listener struct {
	id      string
	ch      chan EventData
	once    bool
	handler func(EventData)
	closed  bool
	mu      sync.Mutex
}

// NewEmitter creates a new event emitter
func NewEmitter() *Emitter {
	return &Emitter{
		listeners: make(map[EventName]map[string]*listener),
	}
}

// On registers an event handler
func (e *Emitter) On(event EventName, handler func(EventData)) string {
	return e.on(event, handler, false)
}

// Once registers a one-time event handler
func (e *Emitter) Once(event EventName, handler func(EventData)) string {
	return e.on(event, handler, true)
}

// on is the internal method to register handlers
func (e *Emitter) on(event EventName, handler func(EventData), once bool) string {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.listeners[event] == nil {
		e.listeners[event] = make(map[string]*listener)
	}

	id := uuid.New().String()
	l := &listener{
		id:      id,
		ch:      make(chan EventData, 100), // Larger buffer to handle bursts
		once:    once,
		handler: handler,
	}

	e.listeners[event][id] = l

	// Start goroutine to handle events
	go func() {
		for data := range l.ch {
			handler(data)
			if once {
				e.Off(event, id)
				return
			}
		}
	}()

	return id
}

// Off removes an event handler by ID
func (e *Emitter) Off(event EventName, id string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if listeners, ok := e.listeners[event]; ok {
		if l, exists := listeners[id]; exists {
			l.mu.Lock()
			if !l.closed {
				l.closed = true
				close(l.ch)
			}
			l.mu.Unlock()
			delete(listeners, id)
		}
	}
}

// Emit emits an event with data
func (e *Emitter) Emit(event EventName, data EventData) {
	e.mu.RLock()
	listeners := e.listeners[event]
	e.mu.RUnlock()

	for _, l := range listeners {
		l.mu.Lock()
		if !l.closed {
			select {
			case l.ch <- data:
			default:
				// Channel full, skip this listener
			}
		}
		l.mu.Unlock()
	}
}

// WaitFor waits for an event with optional filter and context timeout
func (e *Emitter) WaitFor(ctx context.Context, event EventName, filter FilterFunc) (EventData, error) {
	ch := make(chan EventData, 1)
	var once sync.Once

	var handlerID string
	handlerID = e.On(event, func(data EventData) {
		if filter == nil || filter(data) {
			once.Do(func() {
				ch <- data
			})
		}
	})

	select {
	case data := <-ch:
		e.Off(event, handlerID)
		return data, nil
	case <-ctx.Done():
		e.Off(event, handlerID)
		return nil, fmt.Errorf("timeout waiting for event: %s", event)
	}
}

// WaitForAny waits for any of the specified events
func (e *Emitter) WaitForAny(ctx context.Context, events []EventName) (EventName, EventData, error) {
	ch := make(chan struct {
		event EventName
		data  EventData
	}, 1)

	var handlerIDs []struct {
		event EventName
		id    string
	}

	// Register handlers for all events
	for _, event := range events {
		evt := event // Capture loop variable
		id := e.On(evt, func(data EventData) {
			select {
			case ch <- struct {
				event EventName
				data  EventData
			}{evt, data}:
			default:
			}
		})
		handlerIDs = append(handlerIDs, struct {
			event EventName
			id    string
		}{evt, id})
	}

	// Cleanup function
	cleanup := func() {
		for _, h := range handlerIDs {
			e.Off(h.event, h.id)
		}
	}

	select {
	case result := <-ch:
		cleanup()
		return result.event, result.data, nil
	case <-ctx.Done():
		cleanup()
		return "", nil, fmt.Errorf("timeout waiting for events")
	}
}

// RemoveAllListeners removes all listeners for an event
func (e *Emitter) RemoveAllListeners(event EventName) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if listeners, ok := e.listeners[event]; ok {
		for _, l := range listeners {
			l.mu.Lock()
			if !l.closed {
				l.closed = true
				close(l.ch)
			}
			l.mu.Unlock()
		}
		delete(e.listeners, event)
	}
}

// ListenerCount returns the number of listeners for an event
func (e *Emitter) ListenerCount(event EventName) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if listeners, ok := e.listeners[event]; ok {
		return len(listeners)
	}
	return 0
}
