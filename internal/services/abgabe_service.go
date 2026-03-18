package services

import (
	"context"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
)

type AbgabeRepo interface {
	Create(ctx context.Context, ab *abgabe.Abgabe) error
	FindByID(ctx context.Context, id uint) (*abgabe.Abgabe, error)
	ListByUser(ctx context.Context, userID uint) ([]abgabe.Abgabe, error)
	ListByUserAndDateBefore(ctx context.Context, userID uint, date time.Time) ([]abgabe.Abgabe, error)
	Update(ctx context.Context, ab *abgabe.Abgabe) error
	ListByUserAndDateBeforeFromNow(ctx context.Context, userID uint, now, requestDate time.Time) ([]abgabe.Abgabe, error)
	Delete(ctx context.Context, ab *abgabe.Abgabe) error
}

type AbgabeService struct {
	r AbgabeRepo
}

func (s AbgabeService) ListByBeforeDueDate(ctx context.Context, userID uint, beforeDueDate time.Time) ([]abgabe.Abgabe, error) {
	//TODO implement me
	panic("implement me")
}

func NewAbgabeService(r AbgabeRepo) *AbgabeService {
	return &AbgabeService{r: r}
}

func (s AbgabeService) Create(ctx context.Context, userID uint, in abgabe.CreateAbgabeInput) (*abgabe.Abgabe, error) {
	item := &abgabe.Abgabe{
		Title:          in.Title,
		DueDate:        in.DueDate,
		RiskAssessment: abgabe.Risk(in.RiskAssessment),
		UserID:         userID,
	}
	if err := s.r.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s AbgabeService) Get(ctx context.Context, userID, id uint) (*abgabe.Abgabe, error) {
	item, err := s.r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.UserID != userID {
		return nil, custom_errors.ErrForbidden
	}
	return item, nil
}

func (s AbgabeService) List(ctx context.Context, userID uint) ([]abgabe.Abgabe, error) {
	return s.r.ListByUser(ctx, userID)
}

func (s AbgabeService) Update(ctx context.Context, userID, id uint, in abgabe.UpdateAbgabeInput) (*abgabe.Abgabe, error) {
	item, err := s.r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.UserID != userID {
		return nil, custom_errors.ErrForbidden
	}
	if in.Title != nil {
		item.Title = *in.Title
	}
	if in.DueDate != nil {
		item.DueDate = *in.DueDate
	}
	if in.RiskAssessment != nil {
		item.RiskAssessment = abgabe.Risk(*in.RiskAssessment)
	}

	if err := s.r.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s AbgabeService) Delete(ctx context.Context, userID, id uint) error {
	item, err := s.r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if item.UserID != userID {
		return custom_errors.ErrForbidden
	}
	return s.r.Delete(ctx, item)
}
