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

type stubTaskboardService struct {
	createFn  func(ctx context.Context, in services.CreateTaskboardInput) (*models.Taskboard, error)
	listFn    func(ctx context.Context, issuer, sub string, taskID uuid.UUID) ([]models.Taskboard, error)
	getByIDFn func(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Taskboard, error)
	updateFn  func(ctx context.Context, issuer, sub string, id uuid.UUID, in services.UpdateTaskboardInput) (*models.Taskboard, error)
	deleteFn  func(ctx context.Context, issuer, sub string, id uuid.UUID) error
}

func (s *stubTaskboardService) Create(ctx context.Context, in services.CreateTaskboardInput) (*models.Taskboard, error) {
	return s.createFn(ctx, in)
}
func (s *stubTaskboardService) List(ctx context.Context, issuer, sub string, taskID uuid.UUID) ([]models.Taskboard, error) {
	return s.listFn(ctx, issuer, sub, taskID)
}
func (s *stubTaskboardService) GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Taskboard, error) {
	return s.getByIDFn(ctx, issuer, sub, id)
}
func (s *stubTaskboardService) Update(ctx context.Context, issuer, sub string, id uuid.UUID, in services.UpdateTaskboardInput) (*models.Taskboard, error) {
	return s.updateFn(ctx, issuer, sub, id, in)
}
func (s *stubTaskboardService) Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error {
	return s.deleteFn(ctx, issuer, sub, id)
}

func TestTaskboardHandler_Create_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskboardHandler(&stubTaskboardService{})
	r := gin.New()
	r.POST("/tasks/:taskId/boards", h.Create)

	taskID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/boards", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestTaskboardHandler_Create_Success(t *testing.T) {
	taskID := uuid.New()
	svc := &stubTaskboardService{
		createFn: func(_ context.Context, in services.CreateTaskboardInput) (*models.Taskboard, error) {
			return &models.Taskboard{Title: in.Title}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/tasks/:taskId/boards", h.Create)

	body, _ := json.Marshal(map[string]any{"title": "Board1"})
	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/boards", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTaskboardHandler_Create_NotFound(t *testing.T) {
	taskID := uuid.New()
	svc := &stubTaskboardService{
		createFn: func(_ context.Context, _ services.CreateTaskboardInput) (*models.Taskboard, error) {
			return nil, custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/tasks/:taskId/boards", h.Create)

	body, _ := json.Marshal(map[string]any{"title": "Board1"})
	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/boards", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTaskboardHandler_Create_ServiceError(t *testing.T) {
	taskID := uuid.New()
	svc := &stubTaskboardService{
		createFn: func(_ context.Context, _ services.CreateTaskboardInput) (*models.Taskboard, error) {
			return nil, errors.New("unexpected")
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/tasks/:taskId/boards", h.Create)

	body, _ := json.Marshal(map[string]any{"title": "Board1"})
	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/boards", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestTaskboardHandler_List_Success(t *testing.T) {
	taskID := uuid.New()
	svc := &stubTaskboardService{
		listFn: func(_ context.Context, _, _ string, _ uuid.UUID) ([]models.Taskboard, error) {
			return []models.Taskboard{{Title: "A"}}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/tasks/:taskId/boards", h.List)

	req := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID.String()+"/boards", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestTaskboardHandler_Get_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubTaskboardService{
		getByIDFn: func(_ context.Context, _, _ string, _ uuid.UUID) (*models.Taskboard, error) {
			return &models.Taskboard{ID: id, Title: "B"}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/boards/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/boards/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestTaskboardHandler_Get_NotFound(t *testing.T) {
	id := uuid.New()
	svc := &stubTaskboardService{
		getByIDFn: func(_ context.Context, _, _ string, _ uuid.UUID) (*models.Taskboard, error) {
			return nil, custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/boards/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/boards/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTaskboardHandler_Update_NoFields(t *testing.T) {
	id := uuid.New()
	gin.SetMode(gin.TestMode)
	h := NewTaskboardHandler(&stubTaskboardService{})
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.PUT("/boards/:id", h.Update)

	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest(http.MethodPut, "/boards/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTaskboardHandler_Update_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubTaskboardService{
		updateFn: func(_ context.Context, _, _ string, _ uuid.UUID, _ services.UpdateTaskboardInput) (*models.Taskboard, error) {
			return &models.Taskboard{ID: id, Title: "Updated"}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.PUT("/boards/:id", h.Update)

	body, _ := json.Marshal(map[string]any{"title": "Updated"})
	req := httptest.NewRequest(http.MethodPut, "/boards/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestTaskboardHandler_Delete_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubTaskboardService{
		deleteFn: func(_ context.Context, _, _ string, _ uuid.UUID) error { return nil },
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.DELETE("/boards/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/boards/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestTaskboardHandler_Delete_NotFound(t *testing.T) {
	id := uuid.New()
	svc := &stubTaskboardService{
		deleteFn: func(_ context.Context, _, _ string, _ uuid.UUID) error {
			return custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.DELETE("/boards/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/boards/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
