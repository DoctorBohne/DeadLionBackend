package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/DoctorBohne/DeadLionBackend/internal/requestctx"
	"github.com/DoctorBohne/DeadLionBackend/internal/services"
	"github.com/gin-gonic/gin"
)

type AbgabeService interface {
	Create(ctx context.Context, userID uint, in abgabe.CreateAbgabeInput) (*abgabe.Abgabe, error)
	Get(ctx context.Context, userID, id uint) (*abgabe.Abgabe, error)
	List(ctx context.Context, userID uint) ([]abgabe.Abgabe, error)
	ListByBeforeDueDate(ctx context.Context, userID uint, beforeDueDate time.Time) ([]abgabe.Abgabe, error)
	Update(ctx context.Context, userID, id uint, in abgabe.UpdateAbgabeInput) (*abgabe.Abgabe, error)
}

type AbgabeHandler struct {
	abgaben AbgabeService
	usersvc *services.UserService
}

func NewAbgabeHandler(abgaben AbgabeService, usersvc *services.UserService) *AbgabeHandler {
	return &AbgabeHandler{abgaben: abgaben, usersvc: usersvc}
}

func (h *AbgabeHandler) Create(c *gin.Context) {
	user, ok := h.resolveUser(c)
	if !ok {
		return
	}
	var in abgabe.CreateAbgabeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.abgaben.Create(c.Request.Context(), user.ID, in)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"abgabe": item})
}

func (h *AbgabeHandler) Update(c *gin.Context) {
	user, ok := h.resolveUser(c)
	if !ok {
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var in abgabe.UpdateAbgabeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.abgaben.Update(c.Request.Context(), user.ID, id, in)
	if err != nil {
		handleAbgabeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"abgabe": item})
}

func (h *AbgabeHandler) Get(c *gin.Context) {
	user, ok := h.resolveUser(c)
	if !ok {
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	item, err := h.abgaben.Get(c.Request.Context(), user.ID, id)
	if err != nil {
		handleAbgabeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"abgabe": item})
}

func (h *AbgabeHandler) List(c *gin.Context) {
	user, ok := h.resolveUser(c)
	if !ok {
		return
	}
	items, err := h.abgaben.List(c.Request.Context(), user.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"abgaben": items})
}

func (h *AbgabeHandler) resolveUser(c *gin.Context) (*models.User, bool) {
	claims, ok := requestctx.ClaimsFrom(c.Request.Context())
	if !ok || claims.Subject == "" || claims.Issuer == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, false
	}
	in := services.CreateUserInput{
		Issuer:            claims.Issuer,
		Subject:           claims.Subject,
		Email:             claims.Email,
		Name:              claims.Name,
		GivenName:         claims.GivenName,
		FamilyName:        claims.FamilyName,
		EmailVerified:     claims.EmailVerified,
		PreferredUsername: claims.PreferredUsername,
	}
	user, _, err := h.usersvc.FindOrCreate(c.Request.Context(), in)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, false
	}
	return user, true
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	raw := c.Param(name)
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(parsed), true
}

func handleAbgabeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, custom_errors.ErrNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, custom_errors.ErrForbidden):
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
