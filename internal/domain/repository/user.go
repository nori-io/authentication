package repository

import (
	"context"

	"github.com/nori-plugins/authentication/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, e *entity.User) error
	Get(ctx context.Context, id uint64) (*entity.User, error)
	GetAll(ctx context.Context, offset uint64, limit uint64) ([]entity.User, error)
	Update(ctx context.Context, e *entity.User) error
	Delete(ctx context.Context, id uint64) error
}
