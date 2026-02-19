package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/requestctx"
	"github.com/DoctorBohne/DeadLionBackend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	svc services.TaskService
}

func NewTaskHandler(svc services.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) mustClaims(c *gin.Context) (issuer, sub string, ok bool) {
	claims, ok := requestctx.ClaimsFrom(c.Request.Context())
	if !ok || claims.Subject == "" || claims.Issuer == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", "", false
	}
	return claims.Issuer, claims.Subject, true
}

type createTaskReq struct {
	Title           string  `json:"title" binding:"required"`
	Description     *string `json:"description"`
	Risk            *string `json:"risk"`
	Priority        *string `json:"priority"`
	PriorityRank    *string `json:"priorityRank"`
	IsFinished      *string `json:"isFinished"`
	BoardPool       int     `json:"boardPool" binding:"required"`
	EstimateMinutes int     `json:"estimateMinutes" binding:"required"`
	SpendMinutes    int     `json:"spendMinutes" binding:"required"`
	DueAt           string  `json:"dueAt" binding:"required"` // RFC3339
}

func (h *TaskHandler) Create(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	var req createTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "title must not be empty"})
		return
	}

	dueAt, err := time.Parse(time.RFC3339, req.DueAt)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "dueAt must be RFC3339"})
		return
	}

	item, err := h.svc.Create(c.Request.Context(), services.CreateTaskInput{
		Issuer:          issuer,
		Subject:         sub,
		Title:           title,
		Description:     req.Description,
		Risk:            req.Risk,
		Priority:        req.Priority,
		PriorityRank:    req.PriorityRank,
		IsFinished:      req.IsFinished,
		BoardPool:       req.BoardPool,
		EstimateMinutes: req.EstimateMinutes,
		SpendMinutes:    req.SpendMinutes,
		DueAt:           dueAt,
	})
	if err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		default:
			// validation errors -> 400 is nicer than 500
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"task": item})
}

func (h *TaskHandler) List(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	// optional: ?boardPool=3
	var boardPool *int
	if v := c.Query("boardPool"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "boardPool must be int"})
			return
		}
		boardPool = &n
	}

	items, err := h.svc.List(c.Request.Context(), issuer, sub, boardPool)
	if err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *TaskHandler) Get(c *gin.Context) {
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
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"task": item})
}

type updateTaskReq struct {
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	Risk            *string `json:"risk"`
	Priority        *string `json:"priority"`
	PriorityRank    *string `json:"priorityRank"`
	IsFinished      *string `json:"isFinished"`
	BoardPool       *int    `json:"boardPool"`
	EstimateMinutes *int    `json:"estimateMinutes"`
	SpendMinutes    *int    `json:"spendMinutes"`
	DueAt           *string `json:"dueAt"` // RFC3339
}

func (h *TaskHandler) Update(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title == nil && req.Description == nil && req.Risk == nil && req.Priority == nil &&
		req.PriorityRank == nil && req.IsFinished == nil && req.BoardPool == nil &&
		req.EstimateMinutes == nil && req.SpendMinutes == nil && req.DueAt == nil {
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

	var dueAt *time.Time
	if req.DueAt != nil {
		tm, err := time.Parse(time.RFC3339, *req.DueAt)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "dueAt must be RFC3339"})
			return
		}
		dueAt = &tm
	}

	item, err := h.svc.Update(c.Request.Context(), issuer, sub, id, services.UpdateTaskInput{
		Title:           req.Title,
		Description:     req.Description,
		Risk:            req.Risk,
		Priority:        req.Priority,
		PriorityRank:    req.PriorityRank,
		IsFinished:      req.IsFinished,
		BoardPool:       req.BoardPool,
		EstimateMinutes: req.EstimateMinutes,
		SpendMinutes:    req.SpendMinutes,
		DueAt:           dueAt,
	})
	if err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"task": item})
}

func (h *TaskHandler) Delete(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), issuer, sub, id); err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.Status(http.StatusNoContent)
}
