package handler

import (
	"context"
	"net/http"

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
}
type MeHandler struct {
	usersvc UserService
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
