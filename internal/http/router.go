package http

import (
	"github.com/DoctorBohne/DeadLionBackend/internal/http/handler"
	"github.com/DoctorBohne/DeadLionBackend/internal/repositories/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/repositories/boards"
	task "github.com/DoctorBohne/DeadLionBackend/internal/repositories/deadline_objects"
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
	userService := services.NewUserService(*userRepo)
	meHandler := handler.NewMeHandler(*userService)

	abgabeRepo := abgabe.NewAbgabeRepo(d.DB)
	abgabeService := services.NewAbgabeService(*abgabeRepo)
	abgabeHandler := handler.NewAbgabeHandler(abgabeService, userService)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	v1.Use(d.AuthMiddleWare)
	v1.GET("/me", meHandler.Me)
	v1.POST("/me/onboardingcomplete", meHandler.UpdateOnboardingComplete)

	//task bundle
	taskRepo := task.NewTaskRepo(d.DB)
	taskService := services.NewTaskService(taskRepo, userRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	v1.POST("/tasks", taskHandler.Create)
	v1.GET("/tasks", taskHandler.List)
	v1.GET("/tasks/:id", taskHandler.Get)
	v1.PATCH("/tasks/:id", taskHandler.Update)
	v1.DELETE("/tasks/:id", taskHandler.Delete)

	//subtask bundle
	subtaskRepo := task.NewSubtaskRepo(d.DB)
	subtaskService := services.NewSubtaskService(subtaskRepo, userRepo)
	subtaskHandler := handler.NewSubtaskHandler(subtaskService)

	v1.POST("/tasks/:taskId/subtasks", subtaskHandler.Create)
	v1.GET("/subtasks", subtaskHandler.List) // ?taskId=<uuid>
	v1.GET("/subtasks/:id", subtaskHandler.Get)
	v1.PATCH("/subtasks/:id", subtaskHandler.Update)
	v1.DELETE("/subtasks/:id", subtaskHandler.Delete)

	//userboard bundle
	userboardRepo := boards.NewUserboardRepo(d.DB)
	userboardService := services.NewUserboardService(userboardRepo, userRepo)
	userboardHandler := handler.NewUserboardHandler(userboardService)

	v1.POST("/userboard", userboardHandler.Create)
	v1.GET("/userboard", userboardHandler.List)
	v1.GET("/userboard/:id", userboardHandler.Get)
	v1.PATCH("/userboard/:id", userboardHandler.Update)
	v1.DELETE("/userboard/:id", userboardHandler.Delete)

	// boardpool bundle
	boardPoolRepo := boards.NewBoardPoolRepo(d.DB)
	boardPoolService := services.NewBoardPoolService(boardPoolRepo, userRepo)
	boardPoolHandler := handler.NewBoardPoolHandler(boardPoolService)

	v1.POST("/boards/:boardId/pools", boardPoolHandler.Create)
	v1.GET("/boards/:boardId/pools", boardPoolHandler.List)
	v1.GET("/boards/:boardId/pools/:id", boardPoolHandler.Get)
	v1.PATCH("/boards/:boardId/pools/:id", boardPoolHandler.Update)
	v1.DELETE("/boards/:boardId/pools/:id", boardPoolHandler.Delete)

	//taskboard bundle
	taskboardRepo := boards.NewTaskboardRepo(d.DB)
	taskboardService := services.NewTaskboardService(taskboardRepo, userRepo)
	taskboardHandler := handler.NewTaskboardHandler(taskboardService)

	v1.POST("/tasks/:taskId/taskboards", taskboardHandler.Create)
	v1.GET("/taskboards", taskboardHandler.List)
	v1.GET("/taskboards/:id", taskboardHandler.Get) //?taskId=<uuid>
	v1.PATCH("/taskboards/:id", taskboardHandler.Update)
	v1.DELETE("/taskboards/:id", taskboardHandler.Delete)

	v1.POST("/abgaben", abgabeHandler.Create)
	v1.GET("/abgaben", abgabeHandler.List)
	v1.GET("/abgaben/:id", abgabeHandler.Get)
	v1.PUT("/abgaben/:id", abgabeHandler.Update)
	v1.DELETE("/abgaben/:id", abgabeHandler.Delete)

	//riskcalculation bundle
	riskService := services.NewRiskCalculatorService(abgabeRepo)

	riskHandler := handler.NewRiskHandler(riskService, userRepo)
	v1.GET("/abgaben/risklist", riskHandler.RetrieveRiskList)

	return r
}
