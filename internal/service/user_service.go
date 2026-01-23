package service

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/entity"
)

type UserRepo interface {
	Create(ctx context.Context, user *entity.User) error
	FindBySubAndIss(ctx context.Context, sub string, iss string) (*entity.User, error)
}
type UserService struct {
	userRepo UserRepo
}

func NewUserService(userRepo UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

func (u UserService) FindOrCreate(ctx context.Context, iss, sub, email string) (entity.User, error) {
	//TODO implement me
	panic("implement me")
}
