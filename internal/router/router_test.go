package router_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	basev1 "github.com/tanjed/base-go/api/gen/base/v1"
	"github.com/tanjed/base-go/api/gen/base/v1/basev1connect"
	"github.com/tanjed/base-go/internal/apperror"
	healthhandler "github.com/tanjed/base-go/internal/handler/health"
	userhandler "github.com/tanjed/base-go/internal/handler/user"
	"github.com/tanjed/base-go/internal/health"
	"github.com/tanjed/base-go/internal/router"
	usersvc "github.com/tanjed/base-go/internal/service/user"
)

// fakeUsers stands in for the user service: the router test checks transport only.
type fakeUsers struct{ byID map[string]usersvc.User }

func (f *fakeUsers) Create(_ context.Context, in usersvc.Input) (usersvc.User, error) {
	u := usersvc.User{ID: "u1", Name: in.Name, Email: in.Email, CreatedAt: time.Unix(0, 0)}
	f.byID[u.ID] = u
	return u, nil
}

func (f *fakeUsers) Get(_ context.Context, id string) (usersvc.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return usersvc.User{}, apperror.NotFound("user")
	}
	return u, nil
}

func (f *fakeUsers) List(context.Context) ([]usersvc.User, error) {
	out := make([]usersvc.User, 0, len(f.byID))
	for _, u := range f.byID {
		out = append(out, u)
	}
	return out, nil
}

func newServer(t *testing.T) (http.Handler, string, *http.Client) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	opts, err := router.NewHandlerOptions(log)
	require.NoError(t, err)
	r, err := router.New(router.Params{Log: log, Services: []router.Service{
		router.Mount(basev1connect.NewUserServiceHandler, basev1connect.UserServiceHandler(userhandler.New(&fakeUsers{byID: map[string]usersvc.User{}})), opts),
		router.Mount(basev1connect.NewHealthServiceHandler, basev1connect.HealthServiceHandler(healthhandler.New(health.New(nil), log)), opts),
	}})
	require.NoError(t, err)
	srv := httptest.NewUnstartedServer(r.Handler())
	srv.Config = router.NewHTTPServer(r.Handler(), 0)
	srv.Start()
	t.Cleanup(srv.Close)
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	return r.Handler(), srv.URL, &http.Client{Transport: &http.Transport{Protocols: protocols}}
}

func do(t *testing.T, h http.Handler, method, path, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestREST(t *testing.T) {
	h, _, _ := newServer(t)

	code, body := do(t, h, "POST", "/v1/users", `{"name": "Ada", "email": "ada@example.com"}`)
	require.Equal(t, http.StatusOK, code, body)
	require.Equal(t, "ada@example.com", body["user"].(map[string]any)["email"], "snake_case JSON")

	code, body = do(t, h, "POST", "/v1/users", `{"name": "Ada", "emial": "typo"}`)
	require.Equal(t, http.StatusBadRequest, code, "unknown fields are refused")
	require.Equal(t, "invalid_argument", body["error"])

	code, body = do(t, h, "GET", "/v1/users/nobody", "")
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, map[string]any{"error": "not_found", "message": "user not found"}, body, "the reason survives to REST")

	code, _ = do(t, h, "GET", "/ready", "")
	require.Equal(t, http.StatusOK, code)
	code, _ = do(t, h, "GET", "/nowhere", "")
	require.Equal(t, http.StatusNotFound, code)
}

func TestConnectAndGRPC(t *testing.T) {
	_, url, client := newServer(t)
	ctx := context.Background()

	c := basev1connect.NewUserServiceClient(client, url)
	_, err := c.CreateUser(ctx, &basev1.CreateUserRequest{Name: "Ada", Email: "ada@example.com"})
	require.NoError(t, err)

	g := basev1connect.NewUserServiceClient(client, url, connect.WithGRPC())
	res, err := g.ListUsers(ctx, &basev1.ListUsersRequest{})
	require.NoError(t, err)
	require.Len(t, res.GetUsers(), 1)

	_, err = g.GetUser(ctx, &basev1.GetUserRequest{Id: "nobody"})
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	require.Equal(t, "not_found", apperror.Reason(err), "the reason survives to gRPC")
}
