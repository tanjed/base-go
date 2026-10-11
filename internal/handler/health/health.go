// Package health serves HealthService, the kubelet's liveness and readiness probes.
package health

import (
	"context"
	"log/slog"
	"strings"

	basev1 "github.com/tanjed/base-go/api/gen/base/v1"
	basev1connect "github.com/tanjed/base-go/api/gen/base/v1/basev1connect"
	"github.com/tanjed/base-go/internal/apperror"
	checks "github.com/tanjed/base-go/internal/health"
)

var _ basev1connect.HealthServiceHandler = (*Handler)(nil)

// Handler implements HealthService: the kubelet's liveness and readiness probes.
type Handler struct {
	basev1connect.UnimplementedHealthServiceHandler
	Checker checks.Checker
	Log     *slog.Logger
}

func (s *Handler) Live(context.Context, *basev1.LiveRequest) (*basev1.LiveResponse, error) {
	return &basev1.LiveResponse{}, nil
}

func (s *Handler) Ready(ctx context.Context, _ *basev1.ReadyRequest) (*basev1.ReadyResponse, error) {
	var ok, failed []string
	for _, c := range s.Checker.Check(ctx) {
		if c.Err != nil {
			s.Log.WarnContext(ctx, "health check failed", "component", c.Name, "err", c.Err)
			failed = append(failed, c.Name)
			continue
		}
		ok = append(ok, c.Name)
	}
	if len(failed) > 0 {
		// Names only: the error details stay in the log.
		return nil, apperror.Fail(apperror.KindUnavailable, "unhealthy", "unavailable: %s", strings.Join(failed, ", "))
	}
	return &basev1.ReadyResponse{Components: ok}, nil
}
