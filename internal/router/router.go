// Package router serves the service: one listener, Connect, gRPC, gRPC-Web and REST on one port
// (HTTP/1.1 and cleartext HTTP/2).
//
// A handler module declares its service with Serve (its generated New...ServiceHandler, passed as
// is); fx collects them all into the vanguard_services group, and New builds the listener's handler
// from it. REST paths come from the proto's google.api.http annotations.
package router

import (
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"

	"github.com/tanjed/base-go/internal/apperror"
)

// ServicesGroup is the fx value group every Connect service is provided into.
const ServicesGroup = `group:"vanguard_services"`

// Service is one Connect service.
type Service struct {
	Path    string // e.g. /base.v1.UserService/
	Handler http.Handler
}

// Mounter is a generated New...ServiceHandler: it serves the implementation T at its path.
type Mounter[T any] func(T, ...connect.HandlerOption) (string, http.Handler)

// Serve declares a Connect service from a module: mount is its generated New...ServiceHandler, and
// the implementation T comes from the graph (provide it as T, the generated handler interface).
// The service joins the vanguard_services group.
func Serve[T any](mount Mounter[T]) fx.Option {
	return fx.Provide(fx.Annotate(func(impl T, opts HandlerOptions) Service {
		return Mount(mount, impl, opts)
	}, fx.ResultTags(ServicesGroup)))
}

// Mount builds a Service: mount over impl, with the shared handler options (NewHandlerOptions).
func Mount[T any](mount Mounter[T], impl T, opts HandlerOptions) Service {
	path, h := mount(impl, opts...)
	return Service{Path: path, Handler: h}
}

// handler builds the listener's handler. Each service's RPC path goes to the transcoder as it is
// (Connect, gRPC, gRPC-Web keep their own errors); every other path is REST (the transcoder matches
// the google.api.http annotations, unknown paths are 404) behind apperror.REST.
func handler(services []Service) (http.Handler, error) {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)
	if len(services) == 0 {
		return r, nil
	}

	vs := make([]*vanguard.Service, 0, len(services))
	for _, s := range services {
		vs = append(vs, vanguard.NewService(s.Path, s.Handler))
	}
	tc, err := newTranscoder(vs...)
	if err != nil {
		return nil, err
	}
	for _, s := range services {
		r.Handle(s.Path+"*", tc)
	}
	r.Handle("/*", apperror.REST(tc))
	return r, nil
}
