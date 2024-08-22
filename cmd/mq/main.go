package main

import (
	"context"
	"log/slog"
	"os"

	"cloud.google.com/go/pubsub"

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
	cli, err := pubsub.NewClientWithConfig(context.Background(), os.Getenv("GOOGLE_PROJECT_ID"), &pubsub.ClientConfig{
		EnableOpenTelemetryTracing: true,
	})
	if err != nil {
		panic(err)
	}

	topic := cli.Topic(os.Getenv("PUBSUB_TOPIC_ID"))

	if ok, err := topic.Exists(context.Background()); err != nil {
		panic(err)
	} else if !ok {
		if _, err := cli.CreateTopic(context.Background(), os.Getenv("PUBSUB_TOPIC_ID")); err != nil {
			panic(err)
		}
	}

	subscription := cli.Subscription(os.Getenv("PUBSUB_SUBSCRIPTION_ID"))

	if ok, err := subscription.Exists(context.Background()); err != nil {
		panic(err)
	} else if !ok {
		if _, err := cli.CreateSubscription(context.Background(), os.Getenv("PUBSUB_SUBSCRIPTION_ID"), pubsub.SubscriptionConfig{
			Topic: topic,
		}); err != nil {
			panic(err)
		}
	}

	if err := subscription.Receive(context.Background(), func(ctx context.Context, msg *pubsub.Message) {
		slog.InfoContext(ctx, "received message: "+string(msg.Data))
		msg.Ack()
	}); err != nil {
		panic(err)
	}
}
