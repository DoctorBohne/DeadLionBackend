package handler

import (
	"context"
	"net/http"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/DoctorBohne/DeadLionBackend/internal/requestctx"
	"github.com/gin-gonic/gin"
)

type CreateUserInput struct {
	Issuer            string
	Subject           string
	Email             string
	EmailVerified     bool
	Name              string
	PreferredUsername string
	GivenName         string
	FamilyName        string
}

type UserService interface {
	FindOrCreate(ctx context.Context, in CreateUserInput) (*models.User, bool, error)
	MarkOnboardingComplete(ctx context.Context, issuer, sub string) error
}
type MeHandler struct {
	usersvc UserService
}

func NewMeHandler(usersvc UserService) *MeHandler {
	return &MeHandler{usersvc: usersvc}
}

func (m *MeHandler) Me(c *gin.Context) {
	claims, ok := requestctx.ClaimsFrom(c.Request.Context())
	if !ok || claims.Subject == "" || claims.Issuer == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	in := &CreateUserInput{
		Issuer:            claims.Issuer,
		Subject:           claims.Subject,
		Email:             claims.Email,
		Name:              claims.Name,
		GivenName:         claims.GivenName,
		FamilyName:        claims.FamilyName,
		EmailVerified:     claims.EmailVerified,
		PreferredUsername: claims.PreferredUsername,
	}
	user, isNew, err := m.usersvc.FindOrCreate(c.Request.Context(), *in)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if isNew {
		c.JSON(http.StatusCreated, gin.H{"user": user})
	} else {
		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}

func (m *MeHandler) UpdateOnboardingComplete(c *gin.Context) {
	issuerAny, ok := c.Get("issuer")
	if !ok {
		c.JSON(401, gin.H{"error": "missing issuer"})
		return
	}
	subAny, ok := c.Get("sub")
	if !ok {
		c.JSON(401, gin.H{"error": "missing sub"})
		return
	}

	issuer, _ := issuerAny.(string)
	sub, _ := subAny.(string)
	if issuer == "" || sub == "" {
		c.JSON(401, gin.H{"error": "invalid auth context"})
		return
	}

	err := m.usersvc.MarkOnboardingComplete(c.Request.Context(), issuer, sub)
	if err != nil {
		switch err {
		case custom_errors.ErrAlreadBoarded:
			c.JSON(409, gin.H{"error": "onboarding already complete"})
			return
		case custom_errors.ErrNotFound:
			c.JSON(404, gin.H{"error": "user not found"})
			return
		default:
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{"onboardingComplete": true})
}
