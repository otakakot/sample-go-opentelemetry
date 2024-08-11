package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/otakakot/sample-go-opentelemetry/internal/otelx"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

func Middleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := otelx.ExtraTcaceCtx(req)

		slog.InfoContext(ctx, "middleware start")
		defer slog.InfoContext(ctx, "middleware done")

		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

func TraceHTTP(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		name := os.Getenv("SERVICE_NAME")

		ctx, span := otel.Tracer(name).Start(req.Context(), "api")
		defer span.End()

		slog.InfoContext(ctx, "trace start")
		defer slog.InfoContext(ctx, "trace done")

		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

func TraceGPRC(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	slog.InfoContext(ctx, "trace start")
	defer slog.InfoContext(ctx, "trace done")

	return handler(ctx, req)
}
