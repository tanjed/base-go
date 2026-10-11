// Package ioc assembles the service from the packages' fx modules. Each package owns its Module
// (its providers and lifecycle hooks); this only lists them.
//
// The order is the start order: the telemetry providers (telemetry: built first, so they flush
// last on stop), migrations (db), the unit of work (db/uow), repositories,
// services, the handler modules declaring their services (handler/*), then the router building
// them and opening the listener (router, last). Stop runs in reverse.
package ioc

import (
	"go.uber.org/fx"

	"github.com/tanjed/base-go/internal/cache"
	"github.com/tanjed/base-go/internal/config"
	"github.com/tanjed/base-go/internal/db"
	"github.com/tanjed/base-go/internal/db/uow"
	"github.com/tanjed/base-go/internal/events"
	healthhandler "github.com/tanjed/base-go/internal/handler/health"
	userhandler "github.com/tanjed/base-go/internal/handler/user"
	"github.com/tanjed/base-go/internal/health"
	"github.com/tanjed/base-go/internal/logging"
	"github.com/tanjed/base-go/internal/repository"
	"github.com/tanjed/base-go/internal/router"
	"github.com/tanjed/base-go/internal/service"
	"github.com/tanjed/base-go/internal/telemetry"
)

// Modules is the whole service.
var Modules = fx.Options(
	config.Module,
	telemetry.Module,
	logging.Module,
	db.Module,
	uow.Module,
	health.Module,
	cache.Module,
	events.Module,
	repository.Module,
	service.Module,
	healthhandler.Module,
	userhandler.Module,
	router.Module,
)

// New returns the application.
func New() *fx.App { return fx.New(Modules) }
