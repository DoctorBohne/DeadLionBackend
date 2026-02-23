package services

import (
	"context"
	"errors"
	"strings"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserboardRepo interface {
	Create(ctx context.Context, b *models.Userboard) error
	ListByUserID(ctx context.Context, userID uint) ([]models.Userboard, error)
	GetByID(ctx context.Context, userID uint, id uuid.UUID) (*models.Userboard, error)
	Update(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Userboard, error)
	Delete(ctx context.Context, userID uint, id uuid.UUID) (bool, error)
}

type CreateUserboardInput struct {
	Issuer      string
	Subject     string
	Title       string
	Description *string
}

type UpdateUserboardInput struct {
	Title       *string
	Description *string
}

type UserboardService interface {
	Create(ctx context.Context, in CreateUserboardInput) (*models.Userboard, error)
	List(ctx context.Context, issuer, sub string) ([]models.Userboard, error)
	GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Userboard, error)
	Update(ctx context.Context, issuer, sub string, id uuid.UUID, in UpdateUserboardInput) (*models.Userboard, error)
	Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error
}

type userboardService struct {
	users UserLookup
	repo  UserboardRepo
}

func NewUserboardService(repo UserboardRepo, users UserLookup) UserboardService {
	return &userboardService{
		users: users,
		repo:  repo,
	}
}

func (s *userboardService) userID(ctx context.Context, issuer, sub string) (uint, error) {
	u, err := s.users.FindByIssuerSub(ctx, issuer, sub)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, custom_errors.ErrNotFound
		}
		return 0, err
	}
	return u.ID, nil
}

func (s *userboardService) Create(ctx context.Context, in CreateUserboardInput) (*models.Userboard, error) {
	uid, err := s.userID(ctx, in.Issuer, in.Subject)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, errors.New("title must not be empty")
	}

	b := &models.Userboard{
		UserID:      uid,
		Title:       title,
		Description: in.Description,
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *userboardService) List(ctx context.Context, issuer, sub string) ([]models.Userboard, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}
	return s.repo.ListByUserID(ctx, uid)
}

func (s *userboardService) GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Userboard, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}

	b, err := s.repo.GetByID(ctx, uid, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *userboardService) Update(ctx context.Context, issuer, sub string, id uuid.UUID, in UpdateUserboardInput) (*models.Userboard, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}

	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			return nil, errors.New("title must not be empty")
		}
		updates["title"] = t
	}

	if in.Description != nil {
		updates["description"] = in.Description
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	b, err := s.repo.Update(ctx, uid, id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *userboardService) Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return err
	}

	ok, err := s.repo.Delete(ctx, uid, id)
	if err != nil {
		return err
	}
	if !ok {
		return custom_errors.ErrNotFound
	}
	return nil
}
