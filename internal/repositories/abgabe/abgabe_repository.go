package abgabe

import (
	"context"
	"errors"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"gorm.io/gorm"
)

type Repo struct {
	DB *gorm.DB
}

func NewAbgabeRepo(db *gorm.DB) *Repo {
	return &Repo{DB: db}
}

func (r Repo) Create(ctx context.Context, ab *abgabe.Abgabe) error {
	return r.DB.WithContext(ctx).Create(ab).Error
}

func (r Repo) FindByID(ctx context.Context, id uint) (*abgabe.Abgabe, error) {
	var result abgabe.Abgabe
	err := r.DB.WithContext(ctx).First(&result, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return &result, nil
}

func (r Repo) ListByUser(ctx context.Context, userID uint) ([]abgabe.Abgabe, error) {
	var items []abgabe.Abgabe
	err := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("due_date asc").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r Repo) ListByUserAndDateBefore(ctx context.Context, userID uint, date time.Time) ([]abgabe.Abgabe, error) {
	var items []abgabe.Abgabe
	err := r.DB.WithContext(ctx).
		Where("user_id = ? AND date_before = ?", userID, date).
		Order("due_date asc").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r Repo) Update(ctx context.Context, ab *abgabe.Abgabe) error {
	return r.DB.WithContext(ctx).Save(ab).Error
}

func (r Repo) ListByUserAndDateBeforeFromNow(ctx context.Context, userID uint, now, requestDate time.Time) ([]abgabe.Abgabe, error) {
	var items []abgabe.Abgabe
	err := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("due_date >= ? AND due_date <= ?", now, requestDate).
		Order("due_date asc").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r Repo) Delete(ctx context.Context, ab *abgabe.Abgabe) error {
	return r.DB.WithContext(ctx).Delete(ab).Error
}
