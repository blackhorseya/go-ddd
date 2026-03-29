package logx

import (
	"context"
	"log/slog"
)

// fanoutHandler sends log records to multiple handlers.
type fanoutHandler struct {
	handlers []slog.Handler
}

// NewFanoutHandler creates an slog.Handler that fans out to all given handlers.
func NewFanoutHandler(handlers ...slog.Handler) slog.Handler {
	return &fanoutHandler{handlers: handlers}
}

func (h *fanoutHandler) Enabled(c context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(c, level) {
			return true
		}
	}
	return false
}

func (h *fanoutHandler) Handle(c context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(c, r.Level) {
			if err := handler.Handle(c, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &fanoutHandler{handlers: handlers}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &fanoutHandler{handlers: handlers}
}
