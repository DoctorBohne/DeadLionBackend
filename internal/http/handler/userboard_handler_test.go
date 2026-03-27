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

type stubUserboardService struct {
	createFn  func(ctx context.Context, in services.CreateUserboardInput) (*models.Userboard, error)
	listFn    func(ctx context.Context, issuer, sub string) ([]models.Userboard, error)
	getByIDFn func(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Userboard, error)
	updateFn  func(ctx context.Context, issuer, sub string, id uuid.UUID, in services.UpdateUserboardInput) (*models.Userboard, error)
	deleteFn  func(ctx context.Context, issuer, sub string, id uuid.UUID) error
}

func (s *stubUserboardService) Create(ctx context.Context, in services.CreateUserboardInput) (*models.Userboard, error) {
	return s.createFn(ctx, in)
}
func (s *stubUserboardService) List(ctx context.Context, issuer, sub string) ([]models.Userboard, error) {
	return s.listFn(ctx, issuer, sub)
}
func (s *stubUserboardService) GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Userboard, error) {
	return s.getByIDFn(ctx, issuer, sub, id)
}
func (s *stubUserboardService) Update(ctx context.Context, issuer, sub string, id uuid.UUID, in services.UpdateUserboardInput) (*models.Userboard, error) {
	return s.updateFn(ctx, issuer, sub, id, in)
}
func (s *stubUserboardService) Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error {
	return s.deleteFn(ctx, issuer, sub, id)
}

func TestUserboardHandler_Create_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(&stubUserboardService{})
	r := gin.New()
	r.POST("/userboards", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/userboards", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestUserboardHandler_Create_Success(t *testing.T) {
	svc := &stubUserboardService{
		createFn: func(_ context.Context, in services.CreateUserboardInput) (*models.Userboard, error) {
			return &models.Userboard{Title: in.Title}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/userboards", h.Create)

	body, _ := json.Marshal(map[string]any{"title": "My Board"})
	req := httptest.NewRequest(http.MethodPost, "/userboards", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUserboardHandler_Create_UserNotFound(t *testing.T) {
	svc := &stubUserboardService{
		createFn: func(_ context.Context, _ services.CreateUserboardInput) (*models.Userboard, error) {
			return nil, custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/userboards", h.Create)

	body, _ := json.Marshal(map[string]any{"title": "My Board"})
	req := httptest.NewRequest(http.MethodPost, "/userboards", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUserboardHandler_Create_ServiceError(t *testing.T) {
	svc := &stubUserboardService{
		createFn: func(_ context.Context, _ services.CreateUserboardInput) (*models.Userboard, error) {
			return nil, errors.New("unexpected")
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.POST("/userboards", h.Create)

	body, _ := json.Marshal(map[string]any{"title": "My Board"})
	req := httptest.NewRequest(http.MethodPost, "/userboards", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestUserboardHandler_List_Success(t *testing.T) {
	svc := &stubUserboardService{
		listFn: func(_ context.Context, _, _ string) ([]models.Userboard, error) {
			return []models.Userboard{{Title: "B1"}, {Title: "B2"}}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/userboards", h.List)

	req := httptest.NewRequest(http.MethodGet, "/userboards", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUserboardHandler_List_InternalError(t *testing.T) {
	svc := &stubUserboardService{
		listFn: func(_ context.Context, _, _ string) ([]models.Userboard, error) {
			return nil, errors.New("db error")
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/userboards", h.List)

	req := httptest.NewRequest(http.MethodGet, "/userboards", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestUserboardHandler_Get_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubUserboardService{
		getByIDFn: func(_ context.Context, _, _ string, _ uuid.UUID) (*models.Userboard, error) {
			return &models.Userboard{ID: id, Title: "B"}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/userboards/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/userboards/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUserboardHandler_Get_NotFound(t *testing.T) {
	id := uuid.New()
	svc := &stubUserboardService{
		getByIDFn: func(_ context.Context, _, _ string, _ uuid.UUID) (*models.Userboard, error) {
			return nil, custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/userboards/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/userboards/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUserboardHandler_Update_NoFields(t *testing.T) {
	id := uuid.New()
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(&stubUserboardService{})
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.PUT("/userboards/:id", h.Update)

	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest(http.MethodPut, "/userboards/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUserboardHandler_Update_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubUserboardService{
		updateFn: func(_ context.Context, _, _ string, _ uuid.UUID, _ services.UpdateUserboardInput) (*models.Userboard, error) {
			return &models.Userboard{ID: id, Title: "Updated"}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.PUT("/userboards/:id", h.Update)

	body, _ := json.Marshal(map[string]any{"title": "Updated"})
	req := httptest.NewRequest(http.MethodPut, "/userboards/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUserboardHandler_Delete_Success(t *testing.T) {
	id := uuid.New()
	svc := &stubUserboardService{
		deleteFn: func(_ context.Context, _, _ string, _ uuid.UUID) error { return nil },
	}
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.DELETE("/userboards/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/userboards/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestUserboardHandler_Delete_NotFound(t *testing.T) {
	id := uuid.New()
	svc := &stubUserboardService{
		deleteFn: func(_ context.Context, _, _ string, _ uuid.UUID) error {
			return custom_errors.ErrNotFound
		},
	}
	gin.SetMode(gin.TestMode)
	h := NewUserboardHandler(svc)
	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.DELETE("/userboards/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/userboards/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
