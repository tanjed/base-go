package health

import (
	"log/slog"

	"go.uber.org/fx"

	"github.com/tanjed/base-go/api/gen/base/v1/basev1connect"
	checks "github.com/tanjed/base-go/internal/health"
	"github.com/tanjed/base-go/internal/router"
)

// Module serves HealthService (GET /live, /ready: the kubelet's probes).
var Module = fx.Module("healthhandler",
	fx.Provide(fx.Private, fx.Annotate(New, fx.As(new(basev1connect.HealthServiceHandler)))),
	router.Serve(basev1connect.NewHealthServiceHandler),
)

func New(c checks.Checker, log *slog.Logger) *Handler { return &Handler{Checker: c, Log: log} }
