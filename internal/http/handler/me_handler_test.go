package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/DoctorBohne/DeadLionBackend/internal/requestctx"
	"github.com/DoctorBohne/DeadLionBackend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type stubMeUserService struct {
	findOrCreateFn          func(ctx context.Context, in services.CreateUserInput) (*models.User, bool, error)
	markOnboardingCompleteFn func(ctx context.Context, issuer, sub string) error
}

func (s *stubMeUserService) FindOrCreate(ctx context.Context, in services.CreateUserInput) (*models.User, bool, error) {
	return s.findOrCreateFn(ctx, in)
}

func (s *stubMeUserService) MarkOnboardingComplete(ctx context.Context, issuer, sub string) error {
	return s.markOnboardingCompleteFn(ctx, issuer, sub)
}

func TestMeHandler_Me_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubMeUserService{}
	h := NewMeHandler(svc)

	r := gin.New()
	r.GET("/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMeHandler_Me_ExistingUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubMeUserService{
		findOrCreateFn: func(_ context.Context, _ services.CreateUserInput) (*models.User, bool, error) {
			return &models.User{Model: gorm.Model{ID: 1}}, false, nil
		},
	}
	h := NewMeHandler(svc)

	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMeHandler_Me_NewUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubMeUserService{
		findOrCreateFn: func(_ context.Context, _ services.CreateUserInput) (*models.User, bool, error) {
			return &models.User{Model: gorm.Model{ID: 2}}, true, nil
		},
	}
	h := NewMeHandler(svc)

	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestMeHandler_Me_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubMeUserService{
		findOrCreateFn: func(_ context.Context, _ services.CreateUserInput) (*models.User, bool, error) {
			return nil, false, errors.New("db error")
		},
	}
	h := NewMeHandler(svc)

	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func withIssuerSubMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("issuer", "https://issuer.test")
		c.Set("sub", "user-123")
		c.Next()
	}
}

func TestMeHandler_UpdateOnboardingComplete_MissingContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubMeUserService{}
	h := NewMeHandler(svc)

	r := gin.New()
	r.PUT("/me/onboarding", h.UpdateOnboardingComplete)

	req := httptest.NewRequest(http.MethodPut, "/me/onboarding", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMeHandler_UpdateOnboardingComplete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubMeUserService{
		markOnboardingCompleteFn: func(_ context.Context, _, _ string) error { return nil },
	}
	h := NewMeHandler(svc)

	r := gin.New()
	r.Use(withIssuerSubMiddleware())
	r.PUT("/me/onboarding", h.UpdateOnboardingComplete)

	req := httptest.NewRequest(http.MethodPut, "/me/onboarding", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMeHandler_UpdateOnboardingComplete_AlreadyBoarded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubMeUserService{
		markOnboardingCompleteFn: func(_ context.Context, _, _ string) error {
			return custom_errors.ErrAlreadBoarded
		},
	}
	h := NewMeHandler(svc)

	r := gin.New()
	r.Use(withIssuerSubMiddleware())
	r.PUT("/me/onboarding", h.UpdateOnboardingComplete)

	req := httptest.NewRequest(http.MethodPut, "/me/onboarding", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestMeHandler_UpdateOnboardingComplete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubMeUserService{
		markOnboardingCompleteFn: func(_ context.Context, _, _ string) error {
			return custom_errors.ErrNotFound
		},
	}
	h := NewMeHandler(svc)

	r := gin.New()
	r.Use(withIssuerSubMiddleware())
	r.PUT("/me/onboarding", h.UpdateOnboardingComplete)

	req := httptest.NewRequest(http.MethodPut, "/me/onboarding", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestMeHandler_UpdateOnboardingComplete_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubMeUserService{
		markOnboardingCompleteFn: func(_ context.Context, _, _ string) error {
			return errors.New("unexpected")
		},
	}
	h := NewMeHandler(svc)

	r := gin.New()
	r.Use(withIssuerSubMiddleware())
	r.PUT("/me/onboarding", h.UpdateOnboardingComplete)

	req := httptest.NewRequest(http.MethodPut, "/me/onboarding", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// withClaimsCtxMiddleware sets claims via requestctx
func withClaimsCtxMiddleware(claims requestctx.Claims) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := requestctx.WithClaims(c.Request.Context(), claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
