package user

import (
	"go.uber.org/fx"

	"github.com/tanjed/base-go/api/gen/base/v1/basev1connect"
	"github.com/tanjed/base-go/internal/router"
	usersvc "github.com/tanjed/base-go/internal/service/user"
)

// Module binds the user service to the narrow interface UserService declares, and serves it.
var Module = fx.Module("handler.user",
	fx.Provide(
		fx.Private,
		func(s *usersvc.Service) Users { return s },
		fx.Annotate(New, fx.As(new(basev1connect.UserServiceHandler))),
	),
	router.Serve(basev1connect.NewUserServiceHandler),
)

func New(users Users) *Handler { return &Handler{Users: users} }
