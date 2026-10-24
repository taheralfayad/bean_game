package gamevariant

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
	Slug               string `json:"slug" binding:"required"`
	Name               string `json:"name" binding:"required"`
	MinPlayers         uint16 `json:"min_players" binding:"required"`
	MaxPlayers         uint16 `json:"max_players" binding:"required"`
	DeckId             uint16 `json:"deck_id" binding:"required"`
	EndOfTurnDrawCount uint16 `json:"end_of_turn_draw_count" binding:"required"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Register(r gin.IRouter) {
	r.GET("/admin/game-variants", h.list)
	r.POST("/admin/game-variants", h.create)
	r.DELETE("/admin/game-variants/:id", h.delete)
}

func (h *Handler) list(c *gin.Context) {
	variants, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if variants == nil {
		variants = []GameVariant{}
	}

	c.JSON(http.StatusOK, variants)
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug, name, min_players, max_players, deck_id and end_of_turn_draw_count are required"})
		return
	}

	created, err := h.service.Create(c.Request.Context(), GameVariant{
		Slug:               req.Slug,
		Name:               req.Name,
		MinPlayers:         req.MinPlayers,
		MaxPlayers:         req.MaxPlayers,
		DeckId:             req.DeckId,
		EndOfTurnDrawCount: req.EndOfTurnDrawCount,
	})
	if err != nil {
		h.writeCreateError(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 16)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game variant id"})
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
		errors.Is(err, ErrInvalidName),
		errors.Is(err, ErrInvalidPlayers),
		errors.Is(err, ErrInvalidDrawCount):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		c.JSON(http.StatusConflict, gin.H{"error": "game variant slug already exists"})
	case errors.As(err, &pgErr) && pgErr.Code == "23503":
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unknown deck id"})
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
		c.JSON(http.StatusConflict, gin.H{"error": "game variant is in use"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
