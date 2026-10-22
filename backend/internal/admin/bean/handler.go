package bean

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handler struct {
	service *Service
}

type createRequest struct {
	Name  string `json:"name" binding:"required"`
	Image string `json:"image" binding:"required"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/admin/beans", h.create)
	r.DELETE("/admin/beans/:id", h.delete)
}

func (h *Handler) create(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.service.maxBytes*4/3+(1<<20))

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": ErrImageTooLarge.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and image are required"})
		return
	}

	image, err := decodeImage(req.Image)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image must be base64"})
		return
	}

	created, err := h.service.Create(
		c.Request.Context(),
		req.Name,
		Image{Reader: bytes.NewReader(image), Size: int64(len(image))},
	)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 16)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bean id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint16(id)); err != nil {
		h.writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	var pgErr *pgconn.PgError

	switch {
	case errors.Is(err, ErrEmptyImage):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.Is(err, ErrImageTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
	case errors.Is(err, ErrUnsupportedImage):
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		c.JSON(http.StatusConflict, gin.H{"error": "bean name already exists"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func decodeImage(s string) ([]byte, error) {
	if _, data, ok := strings.Cut(s, ","); ok && strings.HasPrefix(s, "data:") {
		s = data
	}

	data, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return data, nil
	}
	return base64.RawStdEncoding.DecodeString(s)
}
