// Package logging provides a structured slog logger that correlates every log
// line for a single HTTP request via its request ID.
package logging

import (
	"context"
	"io"
	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
)

// contextHandler enriches every record with the request ID stored in the
// context (by chi's middleware.RequestID), so all logs emitted while serving a
// request can be correlated.
type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := middleware.GetReqID(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs and WithGroup are overridden so the wrapper survives derived loggers
// (otherwise the embedded handler would be returned and the enrichment lost).
func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{h.Handler.WithGroup(name)}
}

// New returns a JSON slog.Logger that attaches the request ID (when present in
// the context) to every log line.
func New(w io.Writer) *slog.Logger {
	base := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(contextHandler{base})
}
