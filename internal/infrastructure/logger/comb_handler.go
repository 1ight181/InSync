package logger

import (
	"context"
	"log/slog"

	"go.uber.org/multierr"
)

type CombinedHandler struct {
	handlerList []slog.Handler
}

func NewCombinedHandler(handlers []slog.Handler) CombinedHandler {
	return CombinedHandler{handlerList: handlers}
}

func (multiHandler CombinedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, singleHandler := range multiHandler.handlerList {
		if singleHandler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (multiHandler CombinedHandler) Handle(ctx context.Context, record slog.Record) error {
	var resultErr error
	for _, singleHandler := range multiHandler.handlerList {
		if err := singleHandler.Handle(ctx, record); err != nil {
			resultErr = multierr.Append(resultErr, err)
		}
	}
	return resultErr

}

func (multiHandler CombinedHandler) WithAttrs(attributes []slog.Attr) slog.Handler {

	var nextHandlers []slog.Handler

	for _, singleHandler := range multiHandler.handlerList {
		nextHandlers = append(
			nextHandlers,
			singleHandler.WithAttrs(attributes),
		)
	}

	return CombinedHandler{nextHandlers}
}

func (multiHandler CombinedHandler) WithGroup(groupName string) slog.Handler {

	var nextHandlers []slog.Handler

	for _, singleHandler := range multiHandler.handlerList {
		nextHandlers = append(
			nextHandlers,
			singleHandler.WithGroup(groupName),
		)
	}

	return CombinedHandler{nextHandlers}
}
