// Package user is the users table. Every query runs on the unit of work's querier for the
// context: the transaction inside a unit of work, the read replica outside one.
//
// There is no migration for it in this template; a real service adds one, e.g.
//
//	CREATE TABLE users (
//	  id         uuid PRIMARY KEY,
//	  name       text NOT NULL,
//	  email      text NOT NULL UNIQUE,
//	  created_at timestamptz NOT NULL DEFAULT now()
//	);
package user

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/tanjed/base-go/internal/apperror"
	"github.com/tanjed/base-go/internal/db/uow"
)

// Repo is what services use from this repository.
type Repo interface {
	// Create stores a user and returns it as stored; apperror.ErrDuplicate if the email is taken.
	Create(ctx context.Context, u User) (User, error)
	// Get returns a user; apperror.ErrNotFound if there is none.
	Get(ctx context.Context, id string) (User, error)
	List(ctx context.Context) ([]User, error)
}

var _ Repo = (*Repository)(nil)

// User is a row of users.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository struct{ db uow.Querier }

func (r *Repository) Create(ctx context.Context, u User) (User, error) {
	err := r.db.Q(ctx).QueryRow(ctx, `INSERT INTO users (id, name, email) VALUES ($1, $2, $3) RETURNING created_at`,
		u.ID, u.Name, u.Email).Scan(&u.CreatedAt)
	return u, apperror.FromPG(err)
}

func (r *Repository) Get(ctx context.Context, id string) (User, error) {
	rows, err := r.db.Q(ctx).Query(ctx, `SELECT id::text, name, email, created_at FROM users WHERE id = $1`, id)
	if err != nil {
		return User{}, err
	}
	u, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[User])
	return u, apperror.FromPG(err)
}

func (r *Repository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.Q(ctx).Query(ctx, `SELECT id::text, name, email, created_at FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[User])
}
