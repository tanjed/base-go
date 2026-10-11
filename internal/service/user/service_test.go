package user

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/tanjed/base-go/internal/apperror"
	"github.com/tanjed/base-go/internal/cache"
	"github.com/tanjed/base-go/internal/db/uow"
	"github.com/tanjed/base-go/internal/events"
	userrepo "github.com/tanjed/base-go/internal/repository/user"
)

// The service's rules without Postgres: a unit of work that just runs fn, an in-memory repository.

type fakeUoW struct{}

func (fakeUoW) UnitOfWork(ctx context.Context, _ time.Duration, fn func(context.Context) error) error {
	return fn(ctx)
}

var _ uow.Transactor = fakeUoW{}

type fakeRepo struct {
	mu    sync.Mutex
	byID  map[string]User
	reads int
}

var _ userrepo.Repo = (*fakeRepo)(nil)

func (r *fakeRepo) Create(_ context.Context, u User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, x := range r.byID {
		if x.Email == u.Email {
			return User{}, apperror.ErrDuplicate
		}
	}
	u.CreatedAt = time.Now()
	r.byID[u.ID] = u
	return u, nil
}

func (r *fakeRepo) Get(_ context.Context, id string) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reads++
	u, ok := r.byID[id]
	if !ok {
		return User{}, apperror.ErrNotFound
	}
	return u, nil
}

func (r *fakeRepo) List(context.Context) ([]User, error) { return nil, nil }

type fakeEvents struct{ names []string }

func (f *fakeEvents) Publish(_ context.Context, event, _ string, _ map[string]any) {
	f.names = append(f.names, event)
}

var _ events.Publisher = (*fakeEvents)(nil)

func newService(t *testing.T) (*Service, *fakeRepo, *fakeEvents) {
	t.Helper()
	repo, ev := &fakeRepo{byID: map[string]User{}}, &fakeEvents{}
	c := &cache.Redis{C: redis.NewClient(&redis.Options{Addr: miniredis.RunT(t).Addr()}), Prefix: "test"}
	return New(slog.New(slog.NewTextHandler(io.Discard, nil)), fakeUoW{}, c, ev, repo), repo, ev
}

func requireCode(t *testing.T, err error, code string) {
	t.Helper()
	e, ok := apperror.As(err)
	require.True(t, ok, "want an application error, got %v", err)
	require.Equal(t, code, e.Code)
}

func TestCreate(t *testing.T) {
	s, _, ev := newService(t)
	ctx := context.Background()

	u, err := s.Create(ctx, Input{Name: "  Ada ", Email: "ADA@example.com"})
	require.NoError(t, err)
	require.Equal(t, "Ada", u.Name)
	require.Equal(t, "ada@example.com", u.Email)
	require.Equal(t, []string{events.UserCreated}, ev.names)

	_, err = s.Create(ctx, Input{Name: "Ada", Email: "ada@example.com"})
	requireCode(t, err, "email_taken")
	_, err = s.Create(ctx, Input{Name: "", Email: "x@y"})
	requireCode(t, err, "invalid_name")
	_, err = s.Create(ctx, Input{Name: "Ada", Email: "nope"})
	requireCode(t, err, "invalid_email")
}

// Get reads through the cache: the second read does not reach the repository.
func TestGet_Cache(t *testing.T) {
	s, repo, _ := newService(t)
	ctx := context.Background()
	u, err := s.Create(ctx, Input{Name: "Ada", Email: "ada@example.com"})
	require.NoError(t, err)

	got, err := s.Get(ctx, u.ID)
	require.NoError(t, err)
	require.Equal(t, u.ID, got.ID)
	_, err = s.Get(ctx, u.ID)
	require.NoError(t, err)
	require.Equal(t, 1, repo.reads)

	_, err = s.Get(ctx, "not-a-uuid")
	requireCode(t, err, "not_found")
	_, err = s.Get(ctx, "00000000-0000-0000-0000-000000000000")
	requireCode(t, err, "not_found")
}
