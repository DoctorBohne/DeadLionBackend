package handlers

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/entity"
	"github.com/DoctorBohne/DeadLionBackend/internal/service"
	"github.com/gin-gonic/gin"
)

type UserService interface {
	FindOrCreate(ctx context.Context, iss, sub, email string) (entity.User, error)
}

type MeHandler struct {
	userSvc UserService
}

func NewUserHandler(userSvc service.UserService) *MeHandler {
	return &MeHandler{userSvc: userSvc}
}

func (h *MeHandler) Me(c *gin.Context) {

}
