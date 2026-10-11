// Package events publishes domain events to one Kafka topic (APP_KAFKA_TOPIC) in a common
// envelope: { id, event, occurred_at, source, data }, keyed by the given key (e.g. an entity id).
package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Event names: one constant per event this service publishes.
const (
	UserCreated = "USER_CREATED"
)

// Publisher is called after the database commit. Publishing is best effort: a failure is
// logged and never fails the request (there is no outbox yet).
type Publisher interface {
	Publish(ctx context.Context, event, key string, data map[string]any)
}

var (
	_ Publisher = (*Kafka)(nil)
	_ Publisher = Log{}
)

type envelope struct {
	ID         string         `json:"id"`
	Event      string         `json:"event"`
	OccurredAt time.Time      `json:"occurred_at"`
	Source     string         `json:"source"`
	Data       map[string]any `json:"data"`
}

func encode(source, event string, data map[string]any) []byte {
	b, _ := json.Marshal(envelope{ID: uuid.NewString(), Event: event, OccurredAt: time.Now().UTC(), Source: source, Data: data})
	return b
}

// Kafka publishes with acks=all and an idempotent producer.
type Kafka struct {
	client *kgo.Client
	topic  string
	source string
	log    *slog.Logger
}

func (k *Kafka) Publish(ctx context.Context, event, key string, data map[string]any) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	rec := &kgo.Record{Key: []byte(key), Value: encode(k.source, event, data), Headers: []kgo.RecordHeader{{Key: "event", Value: []byte(event)}}}
	if err := k.client.ProduceSync(ctx, rec).FirstErr(); err != nil {
		k.log.ErrorContext(ctx, "event not published", "event", event, "key", key, "err", err)
	}
}

func (k *Kafka) Close() { k.client.Close() }

// Log is the publisher when no brokers are configured: events are only logged.
type Log struct{ Logger *slog.Logger }

func (l Log) Publish(_ context.Context, event, key string, data map[string]any) {
	l.Logger.Info("event (no Kafka configured)", "event", event, "key", key, "data", data)
}
