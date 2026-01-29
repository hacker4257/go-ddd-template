package mysql

import (
	"context"
	"database/sql"
	"errors"

	driver "github.com/go-sql-driver/mysql"

	"github.com/hacker4257/go-ddd-template/internal/domain/user"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, name, email string) (user.User, error) {
	ex := getExecer(r.db, ctx)

	const q = `INSERT INTO users (name, email) VALUES (?, ?)`

	res, err := ex.ExecContext(ctx, q, name, email)
	if err != nil {
		var me *driver.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return user.User{}, user.ErrEmailExists
		}
		return user.User{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return user.User{}, err
	}

	return r.GetByID(ctx, uint64(id))
}

func (r *UserRepo) GetByID(ctx context.Context, id uint64) (user.User, error) {
	ex := getExecer(r.db, ctx)

	const q = `SELECT id, name, email, created_at FROM users WHERE id = ? LIMIT 1`

	var u user.User
	err := ex.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, err
	}

	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (user.User, error) {
	ex := getExecer(r.db, ctx)

	const q = `SELECT id, name, email, created_at FROM users WHERE email = ? LIMIT 1`

	var u user.User
	err := ex.QueryRowContext(ctx, q, email).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, err
	}

	return u, nil
}
