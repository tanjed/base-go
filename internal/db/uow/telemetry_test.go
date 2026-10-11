package uow_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/tanjed/base-go/internal/db"
	"github.com/tanjed/base-go/internal/db/uow"
	"github.com/tanjed/base-go/internal/telemetry/telemetrytest"
	"github.com/tanjed/base-go/internal/testdb"
)

// With a traced pool, every query in a unit of work is a child span of the caller's span.
func TestTelemetry_QueriesAreChildSpans(t *testing.T) {
	testdb.Reset(t, pool)
	rec := telemetrytest.Record(t)
	traced, err := db.Open(context.Background(), pool.Config().ConnString())
	require.NoError(t, err)
	t.Cleanup(traced.Close)
	u := uow.New(db.WritePool{Pool: traced}, db.ReadPool{Pool: traced}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	ctx, parent := otel.Tracer("test").Start(context.Background(), "rpc")
	require.NoError(t, u.UnitOfWork(ctx, time.Second, func(ctx context.Context) error {
		_, err := u.Q(ctx).Exec(ctx, `INSERT INTO companies (id, name) VALUES ($1, 'acme')`, uuid.New())
		return err
	}))
	parent.End()

	var insert sdktrace.ReadOnlySpan
	for _, s := range rec.Ended() {
		if strings.Contains(s.Name(), "INSERT") {
			insert = s
		}
	}
	require.NotNil(t, insert, "insert span")
	require.Equal(t, parent.SpanContext().SpanID(), insert.Parent().SpanID())
}
