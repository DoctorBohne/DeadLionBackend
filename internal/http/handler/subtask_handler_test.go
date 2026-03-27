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

type stubSubtaskService struct {
	createFn  func(ctx context.Context, in services.CreateSubtaskInput) (*models.Subtask, error)
	listFn    func(ctx context.Context, issuer, sub string, taskID uuid.UUID) ([]models.Subtask, error)
	getByIDFn func(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Subtask, error)
	updateFn  func(ctx context.Context, issuer, sub string, id uuid.UUID, in services.UpdateSubtaskInput) (*models.Subtask, error)
	deleteFn  func(ctx context.Context, issuer, sub string, id uuid.UUID) error
}

func (s *stubSubtaskService) Create(ctx context.Context, in services.CreateSubtaskInput) (*models.Subtask, error) {
	return s.createFn(ctx, in)
}
func (s *stubSubtaskService) List(ctx context.Context, issuer, sub string, taskID uuid.UUID) ([]models.Subtask, error) {
	return s.listFn(ctx, issuer, sub, taskID)
}
func (s *stubSubtaskService) GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Subtask, error) {
	return s.getByIDFn(ctx, issuer, sub, id)
}
func (s *stubSubtaskService) Update(ctx context.Context, issuer, sub string, id uuid.UUID, in services.UpdateSubtaskInput) (*models.Subtask, error) {
	return s.updateFn(ctx, issuer, sub, id, in)
}
func (s *stubSubtaskService) Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error {
	return s.deleteFn(ctx, issuer, sub, id)
}

func TestSubtaskHandler_Create_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubtaskHandler(&stubSubtaskService{})
	r := gin.New()
	taskID := uuid.New()
	r.POST("/tasks/:taskId/subtasks", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/subtasks", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestSubtaskHandler_Create_Success(t *testing.T) {
	taskID := uuid.New()
	svc := &stubSubtaskService{
		createFn: func(_ context.Context, in services.CreateSubtaskInput) (*models.Subtask, error) {
			return &models.Subtask{Title: in.Title}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewSubtaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/tasks/:taskId/subtasks", h.Create)

	body, _ := json.Marshal(map[string]any{"title": "Sub1", "boardPool": 1})
	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/subtasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSubtaskHandler_Create_NotFound(t *testing.T) {
	taskID := uuid.New()
	svc := &stubSubtaskService{
		createFn: func(_ context.Context, _ services.CreateSubtaskInput) (*models.Subtask, error) {
			return nil, custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewSubtaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/tasks/:taskId/subtasks", h.Create)

	body, _ := json.Marshal(map[string]any{"title": "Sub1", "boardPool": 1})
	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/subtasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSubtaskHandler_List_Success(t *testing.T) {
	taskID := uuid.New()
	svc := &stubSubtaskService{
		listFn: func(_ context.Context, _, _ string, _ uuid.UUID) ([]models.Subtask, error) {
			return []models.Subtask{{Title: "S1"}}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewSubtaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/tasks/:taskId/subtasks", h.List)

	req := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID.String()+"/subtasks", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSubtaskHandler_List_InternalError(t *testing.T) {
	taskID := uuid.New()
	svc := &stubSubtaskService{
		listFn: func(_ context.Context, _, _ string, _ uuid.UUID) ([]models.Subtask, error) {
			return nil, errors.New("db error")
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewSubtaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/tasks/:taskId/subtasks", h.List)

	req := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID.String()+"/subtasks", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestSubtaskHandler_Get_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubSubtaskService{
		getByIDFn: func(_ context.Context, _, _ string, _ uuid.UUID) (*models.Subtask, error) {
			return &models.Subtask{ID: id, Title: "Sub"}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewSubtaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/subtasks/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/subtasks/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSubtaskHandler_Get_NotFound(t *testing.T) {
	id := uuid.New()
	svc := &stubSubtaskService{
		getByIDFn: func(_ context.Context, _, _ string, _ uuid.UUID) (*models.Subtask, error) {
			return nil, custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewSubtaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/subtasks/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/subtasks/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSubtaskHandler_Update_NoFields(t *testing.T) {
	id := uuid.New()
	gin.SetMode(gin.TestMode)
	h := NewSubtaskHandler(&stubSubtaskService{})
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.PUT("/subtasks/:id", h.Update)

	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest(http.MethodPut, "/subtasks/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSubtaskHandler_Update_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubSubtaskService{
		updateFn: func(_ context.Context, _, _ string, _ uuid.UUID, _ services.UpdateSubtaskInput) (*models.Subtask, error) {
			return &models.Subtask{ID: id, Title: "Updated"}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewSubtaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.PUT("/subtasks/:id", h.Update)

	body, _ := json.Marshal(map[string]any{"title": "Updated"})
	req := httptest.NewRequest(http.MethodPut, "/subtasks/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSubtaskHandler_Delete_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubSubtaskService{
		deleteFn: func(_ context.Context, _, _ string, _ uuid.UUID) error { return nil },
	}
	gin.SetMode(gin.TestMode)
	h := NewSubtaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.DELETE("/subtasks/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/subtasks/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestSubtaskHandler_Delete_NotFound(t *testing.T) {
	id := uuid.New()
	svc := &stubSubtaskService{
		deleteFn: func(_ context.Context, _, _ string, _ uuid.UUID) error {
			return custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewSubtaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.DELETE("/subtasks/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/subtasks/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
