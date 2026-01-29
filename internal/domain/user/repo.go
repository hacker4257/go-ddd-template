package user

import "context"

type Repo interface {
	Create(ctx context.Context, name, email string) (User, error)
	GetByID(ctx context.Context, id uint64) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
}
