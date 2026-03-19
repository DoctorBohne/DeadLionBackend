package services

import (
	"context"
	"errors"
	"testing"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---- stub UserboardRepo ----

type stubUserboardRepo struct {
	createFn      func(ctx context.Context, b *models.Userboard) error
	listByUserFn  func(ctx context.Context, userID uint) ([]models.Userboard, error)
	getByIDFn     func(ctx context.Context, userID uint, id uuid.UUID) (*models.Userboard, error)
	updateFn      func(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Userboard, error)
	deleteFn      func(ctx context.Context, userID uint, id uuid.UUID) (bool, error)
}

func (s *stubUserboardRepo) Create(ctx context.Context, b *models.Userboard) error {
	return s.createFn(ctx, b)
}
func (s *stubUserboardRepo) ListByUserID(ctx context.Context, userID uint) ([]models.Userboard, error) {
	return s.listByUserFn(ctx, userID)
}
func (s *stubUserboardRepo) GetByID(ctx context.Context, userID uint, id uuid.UUID) (*models.Userboard, error) {
	return s.getByIDFn(ctx, userID, id)
}
func (s *stubUserboardRepo) Update(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Userboard, error) {
	return s.updateFn(ctx, userID, id, updates)
}
func (s *stubUserboardRepo) Delete(ctx context.Context, userID uint, id uuid.UUID) (bool, error) {
	return s.deleteFn(ctx, userID, id)
}

// ---- UserboardService tests ----

func TestUserboardService_Create_Success(t *testing.T) {
	repo := &stubUserboardRepo{
		createFn: func(_ context.Context, b *models.Userboard) error { return nil },
	}
	svc := NewUserboardService(repo, &stubUserLookup{user: userWithID(1)})
	board, err := svc.Create(context.Background(), CreateUserboardInput{
		Issuer:  "iss",
		Subject: "sub",
		Title:   "My Board",
	})
	if err != nil || board.Title != "My Board" {
		t.Fatalf("unexpected: %v %v", board, err)
	}
}

func TestUserboardService_Create_UserNotFound(t *testing.T) {
	svc := NewUserboardService(&stubUserboardRepo{}, &stubUserLookup{err: gorm.ErrRecordNotFound})
	_, err := svc.Create(context.Background(), CreateUserboardInput{
		Issuer: "iss", Subject: "sub", Title: "Board",
	})
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserboardService_Create_EmptyTitle(t *testing.T) {
	svc := NewUserboardService(&stubUserboardRepo{}, &stubUserLookup{user: userWithID(1)})
	_, err := svc.Create(context.Background(), CreateUserboardInput{
		Issuer: "iss", Subject: "sub", Title: "   ",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestUserboardService_Create_RepoError(t *testing.T) {
	repo := &stubUserboardRepo{
		createFn: func(_ context.Context, _ *models.Userboard) error { return errors.New("db error") },
	}
	svc := NewUserboardService(repo, &stubUserLookup{user: userWithID(1)})
	_, err := svc.Create(context.Background(), CreateUserboardInput{
		Issuer: "iss", Subject: "sub", Title: "Board",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserboardService_List_Success(t *testing.T) {
	repo := &stubUserboardRepo{
		listByUserFn: func(_ context.Context, _ uint) ([]models.Userboard, error) {
			return []models.Userboard{{Title: "A"}, {Title: "B"}}, nil
		},
	}
	svc := NewUserboardService(repo, &stubUserLookup{user: userWithID(1)})
	items, err := svc.List(context.Background(), "iss", "sub")
	if err != nil || len(items) != 2 {
		t.Fatalf("unexpected: %v %v", items, err)
	}
}

func TestUserboardService_List_UserNotFound(t *testing.T) {
	svc := NewUserboardService(&stubUserboardRepo{}, &stubUserLookup{err: gorm.ErrRecordNotFound})
	_, err := svc.List(context.Background(), "iss", "sub")
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserboardService_GetByID_Success(t *testing.T) {
	id := uuid.New()
	repo := &stubUserboardRepo{
		getByIDFn: func(_ context.Context, _ uint, _ uuid.UUID) (*models.Userboard, error) {
			return &models.Userboard{ID: id, Title: "B"}, nil
		},
	}
	svc := NewUserboardService(repo, &stubUserLookup{user: userWithID(1)})
	board, err := svc.GetByID(context.Background(), "iss", "sub", id)
	if err != nil || board.ID != id {
		t.Fatalf("unexpected: %v %v", board, err)
	}
}

func TestUserboardService_GetByID_NotFound(t *testing.T) {
	id := uuid.New()
	repo := &stubUserboardRepo{
		getByIDFn: func(_ context.Context, _ uint, _ uuid.UUID) (*models.Userboard, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewUserboardService(repo, &stubUserLookup{user: userWithID(1)})
	_, err := svc.GetByID(context.Background(), "iss", "sub", id)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserboardService_Update_Success(t *testing.T) {
	id := uuid.New()
	newTitle := "Updated"
	repo := &stubUserboardRepo{
		updateFn: func(_ context.Context, _ uint, _ uuid.UUID, updates map[string]any) (*models.Userboard, error) {
			return &models.Userboard{ID: id, Title: updates["title"].(string)}, nil
		},
	}
	svc := NewUserboardService(repo, &stubUserLookup{user: userWithID(1)})
	board, err := svc.Update(context.Background(), "iss", "sub", id, UpdateUserboardInput{Title: &newTitle})
	if err != nil || board.Title != "Updated" {
		t.Fatalf("unexpected: %v %v", board, err)
	}
}

func TestUserboardService_Update_NoFields(t *testing.T) {
	id := uuid.New()
	svc := NewUserboardService(&stubUserboardRepo{}, &stubUserLookup{user: userWithID(1)})
	_, err := svc.Update(context.Background(), "iss", "sub", id, UpdateUserboardInput{})
	if err == nil {
		t.Fatal("expected error for empty update")
	}
}

func TestUserboardService_Delete_Success(t *testing.T) {
	id := uuid.New()
	repo := &stubUserboardRepo{
		deleteFn: func(_ context.Context, _ uint, _ uuid.UUID) (bool, error) { return true, nil },
	}
	svc := NewUserboardService(repo, &stubUserLookup{user: userWithID(1)})
	if err := svc.Delete(context.Background(), "iss", "sub", id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserboardService_Delete_NotFound(t *testing.T) {
	id := uuid.New()
	repo := &stubUserboardRepo{
		deleteFn: func(_ context.Context, _ uint, _ uuid.UUID) (bool, error) { return false, nil },
	}
	svc := NewUserboardService(repo, &stubUserLookup{user: userWithID(1)})
	err := svc.Delete(context.Background(), "iss", "sub", id)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
