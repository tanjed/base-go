// Package user is the users' business rules: validation, the unit of work, the cache and events.
package user

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/tanjed/base-go/internal/apperror"
	"github.com/tanjed/base-go/internal/cache"
	"github.com/tanjed/base-go/internal/db/uow"
	"github.com/tanjed/base-go/internal/events"
	userrepo "github.com/tanjed/base-go/internal/repository/user"
)

// User is a user as stored.
type User = userrepo.User

// Input is the body of user create.
type Input struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// cacheTTL is how long a user stays cached after a read.
const cacheTTL = 5 * time.Minute

type Service struct {
	log    *slog.Logger
	uow    uow.Transactor
	cache  cache.Cache
	events events.Publisher
	users  userrepo.Repo
}

// Create stores a new user and publishes USER_CREATED after commit.
func (s *Service) Create(ctx context.Context, in Input) (User, error) {
	in, err := validate(in)
	if err != nil {
		return User{}, err
	}
	var u User
	err = s.uow.UnitOfWork(ctx, uow.Timeout, func(ctx context.Context) error {
		var err error
		u, err = s.users.Create(ctx, User{ID: uuid.NewString(), Name: in.Name, Email: in.Email})
		if errors.Is(err, apperror.ErrDuplicate) {
			return apperror.Fail(apperror.KindConflict, "email_taken", "a user with this email already exists")
		}
		return err
	})
	if err != nil {
		return User{}, err
	}
	s.events.Publish(ctx, events.UserCreated, u.ID, map[string]any{"id": u.ID, "name": u.Name, "email": u.Email})
	return u, nil
}

// Get reads through the cache: a hit skips the database, a miss reads the replica and caches it.
func (s *Service) Get(ctx context.Context, id string) (User, error) {
	if uuid.Validate(id) != nil {
		return User{}, apperror.NotFound("user")
	}
	var u User
	if err := s.cache.Get(ctx, cacheKey(id), &u); err == nil {
		return u, nil
	} else if !errors.Is(err, cache.ErrMiss) {
		s.log.WarnContext(ctx, "cache read failed", "err", err) // degrade to the database
	}
	u, err := s.users.Get(ctx, id)
	if errors.Is(err, apperror.ErrNotFound) {
		return User{}, apperror.NotFound("user")
	}
	if err != nil {
		return User{}, err
	}
	if err := s.cache.Set(ctx, cacheKey(id), u, cacheTTL); err != nil {
		s.log.WarnContext(ctx, "cache write failed", "err", err)
	}
	return u, nil
}

func (s *Service) List(ctx context.Context) ([]User, error) {
	return s.users.List(ctx)
}

func cacheKey(id string) string { return "user:" + id }

// validate trims and checks the input.
func validate(in Input) (Input, error) {
	in.Name, in.Email = strings.TrimSpace(in.Name), strings.ToLower(strings.TrimSpace(in.Email))
	if n := utf8.RuneCountInString(in.Name); n == 0 || n > 100 {
		return in, apperror.Fail(apperror.KindInvalid, "invalid_name", "name must be 1 to 100 characters")
	}
	if !strings.Contains(in.Email, "@") || len(in.Email) > 254 {
		return in, apperror.Fail(apperror.KindInvalid, "invalid_email", "email is not valid")
	}
	return in, nil
}
