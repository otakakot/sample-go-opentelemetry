package slogx

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

var keys = []string{}

type LogHandler struct {
	slog.Handler
}

func New(
	h slog.Handler,
) *LogHandler {
	return &LogHandler{Handler: h}
}

func (lh *LogHandler) Handle(
	ctx context.Context,
	rec slog.Record,
) error {
	span := trace.SpanFromContext(ctx)

	rec.AddAttrs(slog.Attr{Key: "tid", Value: slog.StringValue(span.SpanContext().TraceID().String())})

	rec.AddAttrs(slog.Attr{Key: "sid", Value: slog.StringValue(span.SpanContext().SpanID().String())})

	for _, key := range keys {
		if v := ctx.Value(key); v != nil {
			rec.AddAttrs(slog.Attr{Key: key, Value: slog.AnyValue(v)})
		}
	}

	return lh.Handler.Handle(ctx, rec)
}
