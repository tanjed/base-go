package router

import (
	"log/slog"
	"net/http"

	"go.uber.org/fx"

	"github.com/tanjed/base-go/internal/config"
)

// Module builds the listener's handler from every service in the vanguard_services group and runs
// the listener. Include it last, so it opens after every other start hook.
var Module = fx.Module("router",
	fx.Provide(New),
	fx.Invoke((*Router).Run),
)

// Params is what the router needs; fx fills Services from the vanguard_services group.
type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Shutdown  fx.Shutdowner
	Config    config.Config
	Log       *slog.Logger
	Services  []Service `group:"vanguard_services"`
}

func New(p Params) (*Router, error) {
	h, err := handler(p.Services)
	if err != nil {
		return nil, err
	}
	return &Router{lc: p.Lifecycle, sd: p.Shutdown, cfg: p.Config, log: p.Log, handler: h}, nil
}

// Handler is the listener's handler (tests serve it without opening the configured port).
func (r *Router) Handler() http.Handler { return r.handler }
