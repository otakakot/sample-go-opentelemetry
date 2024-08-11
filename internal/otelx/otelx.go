package otelx

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

func Setup() error {
	endpoint := cmp.Or(os.Getenv("OTLP_ENDPOINT"), "localhost:4317")

	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
		// trace.WithSampler(trace.TraceIDRatioBased(0)),
	)

	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return nil
}

func ExtraTcaceCtx(
	req *http.Request,
) context.Context {
	tc := propagation.TraceContext{}

	return tc.Extract(req.Context(), propagation.HeaderCarrier(req.Header))
}
