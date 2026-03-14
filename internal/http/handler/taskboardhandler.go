package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/requestctx"
	"github.com/DoctorBohne/DeadLionBackend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskboardHandler struct {
	svc services.TaskboardService
}

func NewTaskboardHandler(svc services.TaskboardService) *TaskboardHandler {
	return &TaskboardHandler{svc: svc}
}

func (h *TaskboardHandler) mustClaims(c *gin.Context) (issuer, sub string, ok bool) {
	claims, ok := requestctx.ClaimsFrom(c.Request.Context())
	if !ok || claims.Subject == "" || claims.Issuer == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", "", false
	}
	return claims.Issuer, claims.Subject, true
}

func mustTaskID(c *gin.Context) (uuid.UUID, bool) {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid taskId"})
		return uuid.Nil, false
	}
	return taskID, true
}

type createTaskboardReq struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

func (h *TaskboardHandler) Create(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}
	taskID, ok := mustTaskID(c)
	if !ok {
		return
	}

	var req createTaskboardReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "title must not be empty"})
		return
	}

	board, err := h.svc.Create(c.Request.Context(), services.CreateTaskboardInput{
		Issuer:      issuer,
		Subject:     sub,
		TaskID:      taskID,
		Title:       title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"taskboard": board})
}

func (h *TaskboardHandler) List(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}
	taskID, ok := mustTaskIDQuery(c)
	if !ok {
		return
	}

	items, err := h.svc.List(c.Request.Context(), issuer, sub, taskID)
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *TaskboardHandler) Get(c *gin.Context) {
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
		switch {
		case errors.Is(err, custom_errors.ErrNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "taskboard not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"taskboard": board})
}

type updateTaskboardReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

func (h *TaskboardHandler) Update(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateTaskboardReq
	if err = c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Title == nil && req.Description == nil && req.Status == nil {
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

	board, err := h.svc.Update(c.Request.Context(), issuer, sub, id, services.UpdateTaskboardInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "taskboard not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"taskboard": board})
}

func (h *TaskboardHandler) Delete(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err = h.svc.Delete(c.Request.Context(), issuer, sub, id); err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "taskboard not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.Status(http.StatusNoContent)
}
