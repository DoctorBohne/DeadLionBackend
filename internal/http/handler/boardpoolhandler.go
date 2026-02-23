package handler

import (
	"net/http"
	"strings"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/requestctx"
	"github.com/DoctorBohne/DeadLionBackend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BoardPoolHandler struct {
	svc services.BoardPoolService
}

func NewBoardPoolHandler(svc services.BoardPoolService) *BoardPoolHandler {
	return &BoardPoolHandler{svc: svc}
}

func (h *BoardPoolHandler) mustClaims(c *gin.Context) (issuer, sub string, ok bool) {
	claims, ok := requestctx.ClaimsFrom(c.Request.Context())
	if !ok || claims.Subject == "" || claims.Issuer == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", "", false
	}
	return claims.Issuer, claims.Subject, true
}

func mustBoardID(c *gin.Context) (uuid.UUID, bool) {
	boardID, err := uuid.Parse(c.Param("boardId"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid boardId"})
		return uuid.Nil, false
	}
	return boardID, true
}

type createBoardPoolReq struct {
	Title    string  `json:"title" binding:"required"`
	Color    *string `json:"color"`
	Position *int    `json:"position"`
}

func (h *BoardPoolHandler) Create(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}
	boardID, ok := mustBoardID(c)
	if !ok {
		return
	}

	var req createBoardPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "title must not be empty"})
		return
	}

	item, err := h.svc.Create(c.Request.Context(), services.CreateBoardPoolInput{
		Issuer:   issuer,
		Subject:  sub,
		BoardID:  boardID,
		Title:    title,
		Color:    req.Color,
		Position: req.Position,
	})
	if err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "board not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"boardpool": item})
}

func (h *BoardPoolHandler) List(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}
	boardID, ok := mustBoardID(c)
	if !ok {
		return
	}

	items, err := h.svc.List(c.Request.Context(), issuer, sub, boardID)
	if err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "board not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *BoardPoolHandler) Get(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}
	boardID, ok := mustBoardID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := h.svc.GetByID(c.Request.Context(), issuer, sub, boardID, id)
	if err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "boardpool not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"boardpool": item})
}

type updateBoardPoolReq struct {
	Title    *string `json:"title"`
	Color    *string `json:"color"`
	Position *int    `json:"position"`
}

func (h *BoardPoolHandler) Update(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}
	boardID, ok := mustBoardID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateBoardPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Title == nil && req.Color == nil && req.Position == nil {
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

	item, err := h.svc.Update(c.Request.Context(), issuer, sub, boardID, id, services.UpdateBoardPoolInput{
		Title:    req.Title,
		Color:    req.Color,
		Position: req.Position,
	})
	if err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "boardpool not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"boardpool": item})
}

func (h *BoardPoolHandler) Delete(c *gin.Context) {
	issuer, sub, ok := h.mustClaims(c)
	if !ok {
		return
	}
	boardID, ok := mustBoardID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), issuer, sub, boardID, id); err != nil {
		switch err {
		case custom_errors.ErrNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "boardpool not found"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.Status(http.StatusNoContent)
}
