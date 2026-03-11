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

type SubtaskHandler struct {
	svc services.SubtaskService
}

func NewSubtaskHandler(svc services.SubtaskService) *SubtaskHandler {
	return &SubtaskHandler{svc: svc}
}

func (h *SubtaskHandler) mustClaims(c *gin.Context) (issuer, sub string, ok bool) {
	claims, ok := requestctx.ClaimsFrom(c.Request.Context())
	if !ok || claims.Subject == "" || claims.Issuer == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", "", false
	}
	return claims.Issuer, claims.Subject, true
}

func mustSubtaskID(c *gin.Context) (uuid.UUID, bool) {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid taskId"})
		return uuid.Nil, false
	}
	return taskID, true
}

type createSubtaskReq struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	BoardPool   int     `json:"boardPool" binding:"required"`
}

func (h *SubtaskHandler) Create(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}
	taskID, ok := mustTaskID(c)
	if !ok {
		return
	}

	var req createSubtaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "title must not be empty"})
		return
	}

	item, err := h.svc.Create(c.Request.Context(), services.CreateSubtaskInput{
		Issuer:      issuer,
		Subject:     sub,
		TaskID:      taskID,
		Title:       title,
		Description: req.Description,
		BoardPool:   req.BoardPool,
	})
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"subtask": item})
}

func (h *SubtaskHandler) List(c *gin.Context) {
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

func mustTaskIDQuery(c *gin.Context) (uuid.UUID, bool) {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid taskId"})
		return uuid.Nil, false
	}
	return taskID, true
}

func (h *SubtaskHandler) Get(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := h.svc.GetByID(c.Request.Context(), issuer, sub, id)
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "subtask not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"subtask": item})
}

type updateSubtaskReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	BoardPool   *int    `json:"boardPool"`
}

func (h *SubtaskHandler) Update(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateSubtaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Title == nil && req.Description == nil && req.BoardPool == nil {
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

	item, err := h.svc.Update(c.Request.Context(), issuer, sub, id, services.UpdateSubtaskInput{
		Title:       req.Title,
		Description: req.Description,
		BoardPool:   req.BoardPool,
	})
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "subtask not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"subtask": item})
}

func (h *SubtaskHandler) Delete(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if erro := h.svc.Delete(c.Request.Context(), issuer, sub, id); erro != nil {
		switch {
		case errors.Is(erro, custom_errors.ErrNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "subtask not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": erro.Error()})
			return
		}
	}
	c.Status(http.StatusNoContent)
}
