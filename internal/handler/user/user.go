// Package user serves UserService: it decodes the request, calls the user service and encodes the
// result. Errors are returned untouched (apperror.Interceptor formats them).
package user

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	basev1 "github.com/tanjed/base-go/api/gen/base/v1"
	"github.com/tanjed/base-go/api/gen/base/v1/basev1connect"
	usersvc "github.com/tanjed/base-go/internal/service/user"
)

// Users is what UserService needs from the user service.
type Users interface {
	Create(ctx context.Context, in usersvc.Input) (usersvc.User, error)
	Get(ctx context.Context, id string) (usersvc.User, error)
	List(ctx context.Context) ([]usersvc.User, error)
}

var (
	_ Users                            = (*usersvc.Service)(nil)
	_ basev1connect.UserServiceHandler = (*Handler)(nil)
)

// Handler implements UserService.
type Handler struct {
	basev1connect.UnimplementedUserServiceHandler
	Users Users
}

func (h *Handler) CreateUser(ctx context.Context, req *basev1.CreateUserRequest) (*basev1.CreateUserResponse, error) {
	u, err := h.Users.Create(ctx, usersvc.Input{Name: req.GetName(), Email: req.GetEmail()})
	if err != nil {
		return nil, err
	}
	return &basev1.CreateUserResponse{User: userPB(u)}, nil
}

func (h *Handler) GetUser(ctx context.Context, req *basev1.GetUserRequest) (*basev1.GetUserResponse, error) {
	u, err := h.Users.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &basev1.GetUserResponse{User: userPB(u)}, nil
}

func (h *Handler) ListUsers(ctx context.Context, _ *basev1.ListUsersRequest) (*basev1.ListUsersResponse, error) {
	list, err := h.Users.List(ctx)
	if err != nil {
		return nil, err
	}
	out := &basev1.ListUsersResponse{}
	for _, u := range list {
		out.Users = append(out.Users, userPB(u))
	}
	return out, nil
}

func userPB(u usersvc.User) *basev1.User {
	return &basev1.User{Id: u.ID, Name: u.Name, Email: u.Email, CreateTime: timestamppb.New(u.CreatedAt)}
}
