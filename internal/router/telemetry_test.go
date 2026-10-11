package router_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/tanjed/base-go/api/gen/base/v1/basev1connect"
	userhandler "github.com/tanjed/base-go/internal/handler/user"
	"github.com/tanjed/base-go/internal/router"
	usersvc "github.com/tanjed/base-go/internal/service/user"
	"github.com/tanjed/base-go/internal/telemetry/telemetrytest"
)

// A REST call is one server span, named after its RPC, continuing the caller's trace (the
// gateway's traceparent).
func TestTelemetry_ServerSpanContinuesTrace(t *testing.T) {
	ctx := context.Background()
	rec := telemetrytest.Record(t)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	opts, err := router.NewHandlerOptions(log)
	require.NoError(t, err)
	r, err := router.New(router.Params{Log: log, Services: []router.Service{
		router.Mount(basev1connect.NewUserServiceHandler, basev1connect.UserServiceHandler(userhandler.New(&fakeUsers{byID: map[string]usersvc.User{}})), opts),
	}})
	require.NoError(t, err)

	gateway, span := sdktrace.NewTracerProvider().Tracer("gateway").Start(ctx, "gateway")
	span.End()
	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	propagation.TraceContext{}.Inject(gateway, propagation.HeaderCarrier(req.Header))
	res := httptest.NewRecorder()
	r.Handler().ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	spans := rec.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, trace.SpanKindServer, spans[0].SpanKind())
	require.Equal(t, "base.v1.UserService/ListUsers", spans[0].Name())
	require.Equal(t, trace.SpanContextFromContext(gateway).TraceID(), spans[0].SpanContext().TraceID())
}
