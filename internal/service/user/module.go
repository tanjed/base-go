package user

import (
	"log/slog"

	"go.uber.org/fx"

	"github.com/tanjed/base-go/internal/cache"
	"github.com/tanjed/base-go/internal/db/uow"
	"github.com/tanjed/base-go/internal/events"
	userrepo "github.com/tanjed/base-go/internal/repository/user"
)

// Module provides the user service.
var Module = fx.Module("service.user", fx.Provide(New))

func New(log *slog.Logger, u uow.Transactor, c cache.Cache, pub events.Publisher, users userrepo.Repo) *Service {
	return &Service{log: log, uow: u, cache: c, events: pub, users: users}
}
