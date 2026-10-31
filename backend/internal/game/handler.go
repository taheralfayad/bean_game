package game

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/taheralfayad/bean_game/backend/internal/gamevariant"
)

type Handler struct {
	service *Service
}

type createResponse struct {
	Game   *Game   `json:"game"`
	Player *Player `json:"player"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/games", h.create)
	r.POST("/games/:id/join", h.join)
	r.POST("/games/:id/start", h.start)
}

func (h *Handler) create(c *gin.Context) {
	game, player, err := h.service.CreateGame(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, createResponse{Game: game, Player: player})
}

func (h *Handler) join(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	player, err := h.service.JoinGame(c.Request.Context(), id)
	if err != nil {
		h.writeJoinError(c, err)
		return
	}

	c.JSON(http.StatusCreated, player)
}

func (h *Handler) writeJoinError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrGameNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrGameStarted):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func (h *Handler) start(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	game, err := h.service.StartGame(c.Request.Context(), id)
	if err != nil {
		h.writeStartError(c, err)
		return
	}

	c.JSON(http.StatusOK, game)
}

func (h *Handler) writeStartError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrGameNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrGameStarted):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, gamevariant.ErrNoMatch):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
