package user

import (
	"go.uber.org/fx"

	"github.com/tanjed/base-go/internal/db/uow"
)

// Module provides the user repository as Repo.
var Module = fx.Module("repository.user", fx.Provide(
	fx.Annotate(New, fx.As(new(Repo))),
))

func New(q uow.Querier) *Repository { return &Repository{db: q} }
