package engine

import (
	"context"
	"time"
)

// DefaultEngineTimeout is used by all adapters when no override is set.
const DefaultEngineTimeout = 10 * time.Minute

var engineTimeout = DefaultEngineTimeout

// SetEngineTimeout overrides the global timeout applied to all engine commands.
func SetEngineTimeout(d time.Duration) {
	if d > 0 {
		engineTimeout = d
	}
}

// engineContext returns a context that expires after the current engine timeout.
func engineContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), engineTimeout)
}
