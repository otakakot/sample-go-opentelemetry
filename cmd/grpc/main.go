package main

import (
	"cmp"
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/otakakot/sample-go-opentelemetry/internal/middleware"
	"github.com/otakakot/sample-go-opentelemetry/internal/otelx"
	"github.com/otakakot/sample-go-opentelemetry/internal/slogx"
)

func init() {
	slog.Info("start init")
	defer slog.Info("done init")

	log := slog.New(slogx.New(slog.NewJSONHandler(os.Stdout, nil)))

	slog.SetDefault(log)

	if err := otelx.Setup(); err != nil {
		panic(err)
	}
}

func main() {
	port := cmp.Or(os.Getenv("PORT"), "9090")

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.TraceGPRC),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	go func() {
		slog.Info("start server listen")

		server.Serve(listener)
	}()

	<-ctx.Done()

	slog.Info("start server shutdown")

	server.GracefulStop()

	slog.Info("done server shutdown")
}
