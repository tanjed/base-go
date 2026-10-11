// Package service lists the services' fx modules. Each service is its own package under
// internal/service (business rules only): it uses repositories (internal/repository/*) through the
// unit of work (internal/db/uow).
package service

import (
	"go.uber.org/fx"

	"github.com/tanjed/base-go/internal/service/user"
)

// Module is every service.
var Module = fx.Options(
	user.Module,
)
