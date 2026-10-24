package deck

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handler struct {
	service *Service
}

type createRequest struct {
	Slug  string     `json:"slug" binding:"required"`
	Beans []DeckBean `json:"beans" binding:"required,min=1,dive"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/admin/decks", h.create)
	r.DELETE("/admin/decks/:id", h.delete)
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and a non-empty beans list are required"})
		return
	}

	created, err := h.service.Create(c.Request.Context(), req.Slug, req.Beans)
	if err != nil {
		h.writeCreateError(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 16)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deck id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint16(id)); err != nil {
		h.writeDeleteError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) writeCreateError(c *gin.Context, err error) {
	var pgErr *pgconn.PgError

	switch {
	case errors.Is(err, ErrInvalidSlug),
		errors.Is(err, ErrNoBeans),
		errors.Is(err, ErrInvalidCardCount),
		errors.Is(err, ErrDuplicateBean):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		c.JSON(http.StatusConflict, gin.H{"error": "deck slug already exists"})
	case errors.As(err, &pgErr) && pgErr.Code == "23503":
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unknown bean id"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func (h *Handler) writeDeleteError(c *gin.Context, err error) {
	var pgErr *pgconn.PgError

	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.As(err, &pgErr) && pgErr.Code == "23503":
		c.JSON(http.StatusConflict, gin.H{"error": "deck is in use"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
