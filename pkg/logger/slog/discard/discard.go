package discard

import (
	"context"

	"log/slog"
)

// NewDiscardLogger creates a new logger that discards all messages.
func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
}

type discardHandler struct{}

// NewDiscardHandler creates a new handler that discards all messages produced by a Logger.
func NewDiscardHandler() *discardHandler {
	return &discardHandler{}
}

// Handle handles the Record: do nothing and return nil.
func (h *discardHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

// WithAttrs does nothing and returns the old handler.
func (h *discardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup does nothing and returns the old handler.
func (h *discardHandler) WithGroup(_ string) slog.Handler {
	return h
}

// Enabled reports whether the handler handles records at the given level: returns always false.
func (h *discardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}
