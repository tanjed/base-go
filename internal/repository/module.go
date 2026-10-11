// Package repository lists the repository modules. SQL lives only under internal/repository, one
// package per aggregate; the unit of work is internal/db/uow.
package repository

import (
	"go.uber.org/fx"

	"github.com/tanjed/base-go/internal/repository/user"
)

// Module is every repository.
var Module = fx.Options(
	user.Module,
)
