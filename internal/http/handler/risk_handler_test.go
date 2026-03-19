package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type stubRiskService struct {
	fn func(ctx context.Context, userID uint, requestDate time.Time) ([]models.RiskItem, error)
}

func (s *stubRiskService) CalculateRiskList(ctx context.Context, userID uint, requestDate time.Time) ([]models.RiskItem, error) {
	return s.fn(ctx, userID, requestDate)
}

type stubUserLookup struct {
	user *models.User
	err  error
}

func (s *stubUserLookup) FindByIssuerSub(_ context.Context, _, _ string) (*models.User, error) {
	return s.user, s.err
}

func TestRiskHandler_RetrieveRiskList_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRiskHandler(&stubRiskService{}, &stubUserLookup{})

	r := gin.New()
	r.GET("/risk", h.RetrieveRiskList)

	req := httptest.NewRequest(http.MethodGet, "/risk", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRiskHandler_RetrieveRiskList_UserLookupError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRiskHandler(&stubRiskService{}, &stubUserLookup{err: errors.New("db error")})

	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/risk", h.RetrieveRiskList)

	req := httptest.NewRequest(http.MethodGet, "/risk", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestRiskHandler_RetrieveRiskList_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRiskHandler(
		&stubRiskService{fn: func(_ context.Context, _ uint, _ time.Time) ([]models.RiskItem, error) {
			return []models.RiskItem{{ID: 1, Title: "item1"}}, nil
		}},
		&stubUserLookup{user: &models.User{Model: gorm.Model{ID: 5}}},
	)

	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/risk", h.RetrieveRiskList)

	req := httptest.NewRequest(http.MethodGet, "/risk", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRiskHandler_RetrieveRiskList_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRiskHandler(
		&stubRiskService{fn: func(_ context.Context, _ uint, _ time.Time) ([]models.RiskItem, error) {
			return nil, errors.New("calculation failed")
		}},
		&stubUserLookup{user: &models.User{Model: gorm.Model{ID: 5}}},
	)

	r := gin.New()
	r.Use(withClaimsMiddleware())
	r.GET("/risk", h.RetrieveRiskList)

	req := httptest.NewRequest(http.MethodGet, "/risk", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
