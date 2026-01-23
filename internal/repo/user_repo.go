package repo

import (
	"context"
	"database/sql"

	"github.com/DoctorBohne/DeadLionBackend/internal/entity"
)

type UserRepo struct {
	db *sql.DB
}

func (u UserRepo) Create(ctx context.Context, user *entity.User) error {
	result := u.db.
		panic("implement me")
}

func (u UserRepo) FindBySubAndIss(ctx context.Context, sub string, iss string) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}
