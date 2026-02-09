package services

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/http/handler"
)

type AbgabeRepo interface {
	Create(ctx context.Context, ab *abgabe.Abgabe) error
	FindByID(ctx context.Context, id uint) (*abgabe.Abgabe, error)
	ListByUser(ctx context.Context, userID uint) ([]abgabe.Abgabe, error)
	Update(ctx context.Context, ab *abgabe.Abgabe) error
}

type AbgabeService struct {
	r AbgabeRepo
}

func NewAbgabeService(r AbgabeRepo) *AbgabeService {
	return &AbgabeService{r: r}
}

func (s AbgabeService) Create(ctx context.Context, userID uint, in handler.CreateAbgabeInput) (*abgabe.Abgabe, error) {
	item := &abgabe.Abgabe{
		Title:          in.Title,
		DueDate:        in.DueDate,
		RiskAssessment: in.RiskAssessment,
		UserID:         userID,
		ModulID:        in.ModulID,
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

func (s AbgabeService) Update(ctx context.Context, userID, id uint, in handler.UpdateAbgabeInput) (*abgabe.Abgabe, error) {
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
		item.RiskAssessment = *in.RiskAssessment
	}
	if in.ModulID != nil {
		item.ModulID = *in.ModulID
	}
	if err := s.r.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}
