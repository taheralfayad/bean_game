package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handler struct {
	service *Service
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/login", h.login)
	r.POST("/register", h.register)
}

func (h *Handler) register(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	created, err := h.service.Create(c.Request.Context(), req.Username, []byte(req.Password), UserRole)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	loggedIn, err := h.service.Get(c.Request.Context(), req.Username, []byte(req.Password))
	if err != nil {
		if errors.Is(err, ErrInvalidPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidPassword.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, loggedIn)
}
