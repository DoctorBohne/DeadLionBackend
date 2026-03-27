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

// ---- stub TaskService ----

type stubTaskService struct {
	createFn   func(ctx context.Context, in services.CreateTaskInput) (*models.Task, error)
	listFn     func(ctx context.Context, issuer, sub string, boardPool *int) ([]models.Task, error)
	getByIDFn  func(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Task, error)
	updateFn   func(ctx context.Context, issuer, sub string, id uuid.UUID, in services.UpdateTaskInput) (*models.Task, error)
	deleteFn   func(ctx context.Context, issuer, sub string, id uuid.UUID) error
}

func (s *stubTaskService) Create(ctx context.Context, in services.CreateTaskInput) (*models.Task, error) {
	return s.createFn(ctx, in)
}
func (s *stubTaskService) List(ctx context.Context, issuer, sub string, boardPool *int) ([]models.Task, error) {
	return s.listFn(ctx, issuer, sub, boardPool)
}
func (s *stubTaskService) GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Task, error) {
	return s.getByIDFn(ctx, issuer, sub, id)
}
func (s *stubTaskService) Update(ctx context.Context, issuer, sub string, id uuid.UUID, in services.UpdateTaskInput) (*models.Task, error) {
	return s.updateFn(ctx, issuer, sub, id, in)
}
func (s *stubTaskService) Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error {
	return s.deleteFn(ctx, issuer, sub, id)
}

// ---- helpers ----

func newTaskRouter(svc services.TaskService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	h := NewTaskHandler(svc)
	id := uuid.New().String()
	r.POST("/tasks", h.Create)
	r.GET("/tasks", h.List)
	r.GET("/tasks/"+id, h.Get)
	r.PUT("/tasks/"+id, h.Update)
	r.DELETE("/tasks/"+id, h.Delete)
	return r
}

// ---- TaskHandler.Create ----

func TestTaskHandler_Create_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&stubTaskService{})
	r := gin.New()
	r.POST("/tasks", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/tasks", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestTaskHandler_Create_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&stubTaskService{})
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/tasks", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString("bad json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTaskHandler_Create_Success(t *testing.T) {
	svc := &stubTaskService{
		createFn: func(_ context.Context, in services.CreateTaskInput) (*models.Task, error) {
			return &models.Task{ID: uuid.New(), Title: in.Title}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/tasks", h.Create)

	body, _ := json.Marshal(map[string]any{
		"title":           "My Task",
		"boardPool":       1,
		"estimateMinutes": 30,
		"spendMinutes":    5,
		"dueAt":           "2030-01-01T00:00:00Z",
	})
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTaskHandler_Create_UserNotFound(t *testing.T) {
	svc := &stubTaskService{
		createFn: func(_ context.Context, _ services.CreateTaskInput) (*models.Task, error) {
			return nil, custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/tasks", h.Create)

	body, _ := json.Marshal(map[string]any{
		"title":           "X",
		"boardPool":       1,
		"estimateMinutes": 10,
		"spendMinutes":    1,
		"dueAt":           "2030-01-01T00:00:00Z",
	})
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// ---- TaskHandler.List ----

func TestTaskHandler_List_Success(t *testing.T) {
	svc := &stubTaskService{
		listFn: func(_ context.Context, _, _ string, _ *int) ([]models.Task, error) {
			return []models.Task{{Title: "A"}, {Title: "B"}}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/tasks", h.List)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestTaskHandler_List_NotFound(t *testing.T) {
	svc := &stubTaskService{
		listFn: func(_ context.Context, _, _ string, _ *int) ([]models.Task, error) {
			return nil, custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/tasks", h.List)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTaskHandler_List_InternalError(t *testing.T) {
	svc := &stubTaskService{
		listFn: func(_ context.Context, _, _ string, _ *int) ([]models.Task, error) {
			return nil, errors.New("db error")
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/tasks", h.List)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// ---- TaskHandler.Get ----

func TestTaskHandler_Get_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&stubTaskService{})
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/tasks/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/tasks/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTaskHandler_Get_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubTaskService{
		getByIDFn: func(_ context.Context, _, _ string, _ uuid.UUID) (*models.Task, error) {
			return &models.Task{ID: id, Title: "Test"}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/tasks/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/tasks/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestTaskHandler_Get_NotFound(t *testing.T) {
	id := uuid.New()
	svc := &stubTaskService{
		getByIDFn: func(_ context.Context, _, _ string, _ uuid.UUID) (*models.Task, error) {
			return nil, custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/tasks/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/tasks/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// ---- TaskHandler.Update ----

func TestTaskHandler_Update_NoFields(t *testing.T) {
	id := uuid.New()
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&stubTaskService{})
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.PUT("/tasks/:id", h.Update)

	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest(http.MethodPut, "/tasks/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTaskHandler_Update_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubTaskService{
		updateFn: func(_ context.Context, _, _ string, _ uuid.UUID, _ services.UpdateTaskInput) (*models.Task, error) {
			title := "Updated"
			return &models.Task{ID: id, Title: title}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.PUT("/tasks/:id", h.Update)

	title := "Updated"
	body, _ := json.Marshal(map[string]any{"title": title})
	req := httptest.NewRequest(http.MethodPut, "/tasks/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// ---- TaskHandler.Delete ----

func TestTaskHandler_Delete_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubTaskService{
		deleteFn: func(_ context.Context, _, _ string, _ uuid.UUID) error { return nil },
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.DELETE("/tasks/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestTaskHandler_Delete_NotFound(t *testing.T) {
	id := uuid.New()
	svc := &stubTaskService{
		deleteFn: func(_ context.Context, _, _ string, _ uuid.UUID) error {
			return custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.DELETE("/tasks/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
