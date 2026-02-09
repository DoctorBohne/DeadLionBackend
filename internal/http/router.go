package http

import (
	"github.com/DoctorBohne/DeadLionBackend/internal/http/handler"
	"github.com/DoctorBohne/DeadLionBackend/internal/repositories/user"
	"github.com/DoctorBohne/DeadLionBackend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Deps struct {
	AuthMiddleWare gin.HandlerFunc
	DB             *gorm.DB
}

func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// User bundle
	userRepo := user.NewUserRepo(d.DB)
	userService := services.NewUserService(userRepo)
	meHandler := handler.NewMeHandler(userService)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	v1.Use(d.AuthMiddleWare)
	v1.GET("/me", meHandler.Me)
	v1.POST("/me/onboardingcomplete", meHandler.UpdateOnboardingComplete)
	return r
}
