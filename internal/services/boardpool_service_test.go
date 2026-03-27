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

// ---- stub BoardPoolRepo ----

type stubBoardPoolRepo struct {
	createFn        func(ctx context.Context, userID uint, p *models.BoardPool, explicitPosition bool) error
	listByBoardIDFn func(ctx context.Context, userID uint, boardID uuid.UUID) ([]models.BoardPool, error)
	getByIDFn       func(ctx context.Context, userID uint, boardID, id uuid.UUID) (*models.BoardPool, error)
	updateFn        func(ctx context.Context, userID uint, boardID, id uuid.UUID, updates map[string]any) (*models.BoardPool, error)
	deleteFn        func(ctx context.Context, userID uint, boardID, id uuid.UUID) (bool, error)
}

func (s *stubBoardPoolRepo) Create(ctx context.Context, userID uint, p *models.BoardPool, explicitPosition bool) error {
	return s.createFn(ctx, userID, p, explicitPosition)
}
func (s *stubBoardPoolRepo) ListByBoardID(ctx context.Context, userID uint, boardID uuid.UUID) ([]models.BoardPool, error) {
	return s.listByBoardIDFn(ctx, userID, boardID)
}
func (s *stubBoardPoolRepo) GetByID(ctx context.Context, userID uint, boardID, id uuid.UUID) (*models.BoardPool, error) {
	return s.getByIDFn(ctx, userID, boardID, id)
}
func (s *stubBoardPoolRepo) Update(ctx context.Context, userID uint, boardID, id uuid.UUID, updates map[string]any) (*models.BoardPool, error) {
	return s.updateFn(ctx, userID, boardID, id, updates)
}
func (s *stubBoardPoolRepo) Delete(ctx context.Context, userID uint, boardID, id uuid.UUID) (bool, error) {
	return s.deleteFn(ctx, userID, boardID, id)
}

// ---- BoardPoolService tests ----

func TestBoardPoolService_Create_Success(t *testing.T) {
	boardID := uuid.New()
	repo := &stubBoardPoolRepo{
		createFn: func(_ context.Context, _ uint, _ *models.BoardPool, _ bool) error { return nil },
	}
	svc := NewBoardPoolService(repo, &stubUserLookup{user: userWithID(1)})
	pool, err := svc.Create(context.Background(), CreateBoardPoolInput{
		Issuer:  "iss",
		Subject: "sub",
		BoardID: boardID,
		Title:   "Todo",
	})
	if err != nil || pool.Title != "Todo" {
		t.Fatalf("unexpected: %v %v", pool, err)
	}
}

func TestBoardPoolService_Create_UserNotFound(t *testing.T) {
	svc := NewBoardPoolService(&stubBoardPoolRepo{}, &stubUserLookup{err: gorm.ErrRecordNotFound})
	_, err := svc.Create(context.Background(), CreateBoardPoolInput{
		Issuer: "iss", Subject: "sub", Title: "X",
	})
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBoardPoolService_Create_EmptyTitle(t *testing.T) {
	svc := NewBoardPoolService(&stubBoardPoolRepo{}, &stubUserLookup{user: userWithID(1)})
	_, err := svc.Create(context.Background(), CreateBoardPoolInput{
		Issuer: "iss", Subject: "sub", BoardID: uuid.New(), Title: "  ",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestBoardPoolService_Create_InvalidColor(t *testing.T) {
	color := "notacolor"
	svc := NewBoardPoolService(&stubBoardPoolRepo{}, &stubUserLookup{user: userWithID(1)})
	_, err := svc.Create(context.Background(), CreateBoardPoolInput{
		Issuer: "iss", Subject: "sub", BoardID: uuid.New(), Title: "T", Color: &color,
	})
	if err == nil {
		t.Fatal("expected error for invalid color")
	}
}

func TestBoardPoolService_Create_ValidColor(t *testing.T) {
	boardID := uuid.New()
	color := "#ff0000"
	repo := &stubBoardPoolRepo{
		createFn: func(_ context.Context, _ uint, _ *models.BoardPool, _ bool) error { return nil },
	}
	svc := NewBoardPoolService(repo, &stubUserLookup{user: userWithID(1)})
	pool, err := svc.Create(context.Background(), CreateBoardPoolInput{
		Issuer: "iss", Subject: "sub", BoardID: boardID, Title: "Red", Color: &color,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool.Color != "ff0000" {
		t.Fatalf("unexpected color: %s", pool.Color)
	}
}

func TestBoardPoolService_List_Success(t *testing.T) {
	boardID := uuid.New()
	repo := &stubBoardPoolRepo{
		listByBoardIDFn: func(_ context.Context, _ uint, _ uuid.UUID) ([]models.BoardPool, error) {
			return []models.BoardPool{{Title: "A"}, {Title: "B"}}, nil
		},
	}
	svc := NewBoardPoolService(repo, &stubUserLookup{user: userWithID(1)})
	items, err := svc.List(context.Background(), "iss", "sub", boardID)
	if err != nil || len(items) != 2 {
		t.Fatalf("unexpected: %v %v", items, err)
	}
}

func TestBoardPoolService_GetByID_Success(t *testing.T) {
	boardID, id := uuid.New(), uuid.New()
	repo := &stubBoardPoolRepo{
		getByIDFn: func(_ context.Context, _ uint, _, _ uuid.UUID) (*models.BoardPool, error) {
			return &models.BoardPool{ID: id, Title: "P"}, nil
		},
	}
	svc := NewBoardPoolService(repo, &stubUserLookup{user: userWithID(1)})
	pool, err := svc.GetByID(context.Background(), "iss", "sub", boardID, id)
	if err != nil || pool.ID != id {
		t.Fatalf("unexpected: %v %v", pool, err)
	}
}

func TestBoardPoolService_GetByID_NotFound(t *testing.T) {
	boardID, id := uuid.New(), uuid.New()
	repo := &stubBoardPoolRepo{
		getByIDFn: func(_ context.Context, _ uint, _, _ uuid.UUID) (*models.BoardPool, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewBoardPoolService(repo, &stubUserLookup{user: userWithID(1)})
	_, err := svc.GetByID(context.Background(), "iss", "sub", boardID, id)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBoardPoolService_Delete_Success(t *testing.T) {
	boardID, id := uuid.New(), uuid.New()
	repo := &stubBoardPoolRepo{
		deleteFn: func(_ context.Context, _ uint, _, _ uuid.UUID) (bool, error) { return true, nil },
	}
	svc := NewBoardPoolService(repo, &stubUserLookup{user: userWithID(1)})
	if err := svc.Delete(context.Background(), "iss", "sub", boardID, id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBoardPoolService_Delete_NotFound(t *testing.T) {
	boardID, id := uuid.New(), uuid.New()
	repo := &stubBoardPoolRepo{
		deleteFn: func(_ context.Context, _ uint, _, _ uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	svc := NewBoardPoolService(repo, &stubUserLookup{user: userWithID(1)})
	err := svc.Delete(context.Background(), "iss", "sub", boardID, id)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
