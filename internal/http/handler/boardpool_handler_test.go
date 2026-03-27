package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/DoctorBohne/DeadLionBackend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubBoardPoolService struct {
	createFn  func(ctx context.Context, in services.CreateBoardPoolInput) (*models.BoardPool, error)
	listFn    func(ctx context.Context, issuer, sub string, boardID uuid.UUID) ([]models.BoardPool, error)
	getByIDFn func(ctx context.Context, issuer, sub string, boardID, id uuid.UUID) (*models.BoardPool, error)
	updateFn  func(ctx context.Context, issuer, sub string, boardID, id uuid.UUID, in services.UpdateBoardPoolInput) (*models.BoardPool, error)
	deleteFn  func(ctx context.Context, issuer, sub string, boardID, id uuid.UUID) error
}

func (s *stubBoardPoolService) Create(ctx context.Context, in services.CreateBoardPoolInput) (*models.BoardPool, error) {
	return s.createFn(ctx, in)
}
func (s *stubBoardPoolService) List(ctx context.Context, issuer, sub string, boardID uuid.UUID) ([]models.BoardPool, error) {
	return s.listFn(ctx, issuer, sub, boardID)
}
func (s *stubBoardPoolService) GetByID(ctx context.Context, issuer, sub string, boardID, id uuid.UUID) (*models.BoardPool, error) {
	return s.getByIDFn(ctx, issuer, sub, boardID, id)
}
func (s *stubBoardPoolService) Update(ctx context.Context, issuer, sub string, boardID, id uuid.UUID, in services.UpdateBoardPoolInput) (*models.BoardPool, error) {
	return s.updateFn(ctx, issuer, sub, boardID, id, in)
}
func (s *stubBoardPoolService) Delete(ctx context.Context, issuer, sub string, boardID, id uuid.UUID) error {
	return s.deleteFn(ctx, issuer, sub, boardID, id)
}

func TestBoardPoolHandler_Create_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewBoardPoolHandler(&stubBoardPoolService{})
	r := gin.New()
	boardID := uuid.New()
	r.POST("/boards/:boardId/pools", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/boards/"+boardID.String()+"/pools", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBoardPoolHandler_Create_Success(t *testing.T) {
	boardID := uuid.New()
	svc := &stubBoardPoolService{
		createFn: func(_ context.Context, in services.CreateBoardPoolInput) (*models.BoardPool, error) {
			return &models.BoardPool{Title: in.Title}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewBoardPoolHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/boards/:boardId/pools", h.Create)

	body, _ := json.Marshal(map[string]any{"title": "Pool1"})
	req := httptest.NewRequest(http.MethodPost, "/boards/"+boardID.String()+"/pools", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBoardPoolHandler_Create_NotFound(t *testing.T) {
	boardID := uuid.New()
	svc := &stubBoardPoolService{
		createFn: func(_ context.Context, _ services.CreateBoardPoolInput) (*models.BoardPool, error) {
			return nil, custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewBoardPoolHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/boards/:boardId/pools", h.Create)

	body, _ := json.Marshal(map[string]any{"title": "Pool1"})
	req := httptest.NewRequest(http.MethodPost, "/boards/"+boardID.String()+"/pools", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestBoardPoolHandler_Create_ServiceError(t *testing.T) {
	boardID := uuid.New()
	svc := &stubBoardPoolService{
		createFn: func(_ context.Context, _ services.CreateBoardPoolInput) (*models.BoardPool, error) {
			return nil, errors.New("unexpected")
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewBoardPoolHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/boards/:boardId/pools", h.Create)

	body, _ := json.Marshal(map[string]any{"title": "Pool1"})
	req := httptest.NewRequest(http.MethodPost, "/boards/"+boardID.String()+"/pools", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestBoardPoolHandler_List_Success(t *testing.T) {
	boardID := uuid.New()
	svc := &stubBoardPoolService{
		listFn: func(_ context.Context, _, _ string, _ uuid.UUID) ([]models.BoardPool, error) {
			return []models.BoardPool{{Title: "P1"}}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewBoardPoolHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/boards/:boardId/pools", h.List)

	req := httptest.NewRequest(http.MethodGet, "/boards/"+boardID.String()+"/pools", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBoardPoolHandler_Get_Success(t *testing.T) {
	boardID, id := uuid.New(), uuid.New()
	svc := &stubBoardPoolService{
		getByIDFn: func(_ context.Context, _, _ string, _, _ uuid.UUID) (*models.BoardPool, error) {
			return &models.BoardPool{ID: id}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewBoardPoolHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/boards/:boardId/pools/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/boards/"+boardID.String()+"/pools/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBoardPoolHandler_Get_NotFound(t *testing.T) {
	boardID, id := uuid.New(), uuid.New()
	svc := &stubBoardPoolService{
		getByIDFn: func(_ context.Context, _, _ string, _, _ uuid.UUID) (*models.BoardPool, error) {
			return nil, custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewBoardPoolHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/boards/:boardId/pools/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/boards/"+boardID.String()+"/pools/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestBoardPoolHandler_Update_NoFields(t *testing.T) {
	boardID, id := uuid.New(), uuid.New()
	gin.SetMode(gin.TestMode)
	h := NewBoardPoolHandler(&stubBoardPoolService{})
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.PUT("/boards/:boardId/pools/:id", h.Update)

	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest(http.MethodPut, "/boards/"+boardID.String()+"/pools/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBoardPoolHandler_Update_Success(t *testing.T) {
	boardID, id := uuid.New(), uuid.New()
	svc := &stubBoardPoolService{
		updateFn: func(_ context.Context, _, _ string, _, _ uuid.UUID, _ services.UpdateBoardPoolInput) (*models.BoardPool, error) {
			return &models.BoardPool{ID: id, Title: "Updated"}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewBoardPoolHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.PUT("/boards/:boardId/pools/:id", h.Update)

	body, _ := json.Marshal(map[string]any{"title": "Updated"})
	req := httptest.NewRequest(http.MethodPut, "/boards/"+boardID.String()+"/pools/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBoardPoolHandler_Delete_Success(t *testing.T) {
	boardID, id := uuid.New(), uuid.New()
	svc := &stubBoardPoolService{
		deleteFn: func(_ context.Context, _, _ string, _, _ uuid.UUID) error { return nil },
	}
	gin.SetMode(gin.TestMode)
	h := NewBoardPoolHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.DELETE("/boards/:boardId/pools/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/boards/"+boardID.String()+"/pools/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestBoardPoolHandler_Delete_NotFound(t *testing.T) {
	boardID, id := uuid.New(), uuid.New()
	svc := &stubBoardPoolService{
		deleteFn: func(_ context.Context, _, _ string, _, _ uuid.UUID) error {
			return custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewBoardPoolHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.DELETE("/boards/:boardId/pools/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/boards/"+boardID.String()+"/pools/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
