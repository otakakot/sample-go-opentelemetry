package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	port := cmp.Or(os.Getenv("PORT"), "8080")

	conn, err := grpc.NewClient(
		cmp.Or(os.Getenv("GRPC_ENDPOINT"), "localhost:9090"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Warn("failed to close grpc connection. error: " + err.Error())
		}
	}()

	hdl := NewHandler(conn)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", hdl.Health)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           middleware.TraceHTTP(mux),
		ReadHeaderTimeout: 30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	go func() {
		slog.Info("start server listen")

		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	<-ctx.Done()

	slog.Info("start server shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}

	slog.Info("done server shutdown")
}

type Handler struct {
	restclient *http.Client
	grpclient  healthpb.HealthClient
}

func NewHandler(
	conn *grpc.ClientConn,
) *Handler {
	grpcclient := healthpb.NewHealthClient(conn)

	return &Handler{
		restclient: &http.Client{},
		grpclient:  grpcclient,
	}
}

func (hdl *Handler) Health(
	rw http.ResponseWriter,
	req *http.Request,
) {
	slog.InfoContext(req.Context(), "api health check")
	defer slog.InfoContext(req.Context(), "api health check done")

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)

	defer cancel()

	if _, err := hdl.grpclient.Check(ctx, &healthpb.HealthCheckRequest{
		Service: "",
	}); err != nil {
		slog.ErrorContext(req.Context(), "failed to check grpc health. error: "+err.Error())

		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	fmt.Fprintf(rw, "OK")
}
