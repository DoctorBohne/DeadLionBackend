package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/DoctorBohne/DeadLionBackend/internal/requestctx"
	"github.com/DoctorBohne/DeadLionBackend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type stubUserService struct {
	user *models.User
	err  error
}

func (s stubUserService) FindOrCreate(ctx context.Context, in services.CreateUserInput) (*models.User, bool, error) {
	return s.user, false, s.err
}

type stubAbgabeService struct {
	createFn func(ctx context.Context, userID uint, in abgabe.CreateAbgabeInput) (*abgabe.Abgabe, error)
	getFn    func(ctx context.Context, userID, id uint) (*abgabe.Abgabe, error)
	listFn   func(ctx context.Context, userID uint) ([]abgabe.Abgabe, error)
	updateFn func(ctx context.Context, userID, id uint, in abgabe.UpdateAbgabeInput) (*abgabe.Abgabe, error)
}

func (s stubAbgabeService) ListByBeforeDueDate(ctx context.Context, userID uint, beforeDueDate time.Time) ([]abgabe.Abgabe, error) {
	return nil, nil
	//probably needs fixing xd
}

func (s stubAbgabeService) Create(ctx context.Context, userID uint, in abgabe.CreateAbgabeInput) (*abgabe.Abgabe, error) {
	return s.createFn(ctx, userID, in)
}

func (s stubAbgabeService) Get(ctx context.Context, userID, id uint) (*abgabe.Abgabe, error) {
	return s.getFn(ctx, userID, id)
}

func (s stubAbgabeService) List(ctx context.Context, userID uint) ([]abgabe.Abgabe, error) {
	return s.listFn(ctx, userID)
}

func (s stubAbgabeService) Update(ctx context.Context, userID, id uint, in abgabe.UpdateAbgabeInput) (*abgabe.Abgabe, error) {
	return s.updateFn(ctx, userID, id, in)
}

func withClaimsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := requestctx.Claims{
			Issuer:  "https://issuer.test",
			Subject: "user-123",
		}
		ctx := requestctx.WithClaims(c.Request.Context(), claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func TestAbgabeHandlerCreateSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expectedDueDate := time.Date(2025, 6, 2, 9, 30, 0, 0, time.UTC)
	var gotUserID uint
	var gotInput abgabe.CreateAbgabeInput

	abService := stubAbgabeService{
		createFn: func(ctx context.Context, userID uint, in abgabe.CreateAbgabeInput) (*abgabe.Abgabe, error) {
			gotUserID = userID
			gotInput = in
			return &abgabe.Abgabe{Title: in.Title, DueDate: in.DueDate, RiskAssessment: abgabe.Risk(in.RiskAssessment), UserID: userID}, nil
		},
	}
	userService := stubUserService{user: &models.User{Model: gorm.Model{ID: 42}}}

	r := gin.New()
	r.Use(withClaimsMiddleware())
	handler := NewAbgabeHandler(abService, userService)
	r.POST("/api/v1/abgaben", handler.Create)

	payload := map[string]any{
		"title":           "Mathe Aufgaben",
		"due_date":        expectedDueDate.Format(time.RFC3339),
		"risk_assessment": 2,
		"modul_id":        3,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/abgaben", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
	if gotUserID != 42 {
		t.Fatalf("expected user id 42, got %d", gotUserID)
	}
	if gotInput.Title != "Mathe Aufgaben" || !gotInput.DueDate.Equal(expectedDueDate) {
		t.Fatalf("unexpected input data: %#v", gotInput)
	}
}

func TestAbgabeHandlerListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	abService := stubAbgabeService{
		listFn: func(ctx context.Context, userID uint) ([]abgabe.Abgabe, error) {
			return []abgabe.Abgabe{{Title: "A"}, {Title: "B"}}, nil
		},
	}
	userService := stubUserService{user: &models.User{Model: gorm.Model{ID: 7}}}

	r := gin.New()
	r.Use(withClaimsMiddleware())
	handler := NewAbgabeHandler(abService, userService)
	r.GET("/api/v1/abgaben", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/abgaben", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var response struct {
		Abgaben []abgabe.Abgabe `json:"abgaben"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(response.Abgaben) != 2 {
		t.Fatalf("expected 2 abgaben, got %d", len(response.Abgaben))
	}
}

func TestAbgabeHandlerGetUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	abService := stubAbgabeService{
		getFn: func(ctx context.Context, userID, id uint) (*abgabe.Abgabe, error) {
			return &abgabe.Abgabe{}, nil
		},
	}
	userService := stubUserService{user: &models.User{Model: gorm.Model{ID: 1}}}

	r := gin.New()
	handler := NewAbgabeHandler(abService, userService)
	r.GET("/api/v1/abgaben/:id", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/abgaben/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
