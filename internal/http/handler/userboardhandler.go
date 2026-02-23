package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/DoctorBohne/DeadLionBackend/internal/requestctx"
	"github.com/DoctorBohne/DeadLionBackend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateUserboardInput struct {
	Issuer      string
	Subject     string
	Title       string
	Description *string
}

type UpdateUserboardInput struct {
	Title       *string
	Description *string
}

type UserboardService interface {
	Create(ctx context.Context, in CreateUserboardInput) (*models.Userboard, error)
	List(ctx context.Context, issuer, sub string) ([]models.Userboard, error)
	GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Userboard, error)
	Update(ctx context.Context, issuer, sub string, id uuid.UUID, in UpdateUserboardInput) (*models.Userboard, error)
	Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error
}

type UserboardHandler struct {
	svc services.UserboardService
}

func NewUserboardHandler(svc services.UserboardService) *UserboardHandler {
	return &UserboardHandler{svc: svc}
}

func (h *UserboardHandler) mustClaims(c *gin.Context) (issuer, sub string, ok bool) {
	claims, ok := requestctx.ClaimsFrom(c.Request.Context())
	if !ok || claims.Subject == "" || claims.Issuer == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", "", false
	}
	return claims.Issuer, claims.Subject, true
}

type createUserboardReq struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
}

func (h *UserboardHandler) Create(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	var req createUserboardReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "title must not be empty"})
		return
	}

	in := services.CreateUserboardInput{
		Issuer:      issuer,
		Subject:     sub,
		Title:       title,
		Description: req.Description,
	}

	board, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		if err == custom_errors.ErrNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"userboard": board})
}

func (h *UserboardHandler) List(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	items, err := h.svc.List(c.Request.Context(), issuer, sub)
	if err != nil {
		if err == custom_errors.ErrNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *UserboardHandler) Get(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	board, err := h.svc.GetByID(c.Request.Context(), issuer, sub, id)
	if err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "userboard not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"userboard": board})
}

type updateUserboardReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

func (h *UserboardHandler) Update(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateUserboardReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Title == nil && req.Description == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "title must not be empty"})
			return
		}
		req.Title = &t
	}

	board, err := h.svc.Update(c.Request.Context(), issuer, sub, id, services.UpdateUserboardInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "userboard not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"userboard": board})
}

func (h *UserboardHandler) Delete(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.svc.Delete(c.Request.Context(), issuer, sub, id)
	if err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "userboard not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.Status(http.StatusNoContent)
}
