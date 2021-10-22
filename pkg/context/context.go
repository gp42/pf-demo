// Package context contains custom context value ids and context-related functions
package context

import (
	"context"

	"github.com/google/uuid"
)

const (
	// REQ_KEY_ID Request ID context value key
	REQ_KEY_ID = iota
	// CLIENT_IP_KEY_ID Client IP context value key
	CLIENT_IP_KEY_ID = iota
)

// WithRandomRequestID stores random request ID in context values
func WithRandomRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, REQ_KEY_ID, uuid.New().String())
}
