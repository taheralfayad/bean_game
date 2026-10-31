package gameevents

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const heartbeatInterval = 25 * time.Second

type Handler struct {
	broker *Broker
}

func NewHandler(broker *Broker) *Handler { return &Handler{broker: broker} }

var (
	ErrStreamingNotSupported = errors.New("streaming not supported")
	ErrMissingTopic          = errors.New("topic is required")
)

func (h *Handler) Register(r gin.IRouter) {
	r.GET("/gameevents/:topic", h.getGameEvents)
}

func (h *Handler) getGameEvents(c *gin.Context) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrStreamingNotSupported.Error()})
		return
	}

	topic := c.Param("topic")
	if topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrMissingTopic.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	events, unsubscribe := h.broker.Subscribe(topic)
	defer unsubscribe()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			if _, err := io.WriteString(c.Writer, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := writeEvent(c.Writer, event); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeEvent(w io.Writer, event Event) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}

	var buf strings.Builder

	if event.ID != "" {
		fmt.Fprintf(&buf, "id: %s\n", event.ID)
	}

	if event.Type != "" {
		fmt.Fprintf(&buf, "event: %s\n", event.Type)
	}

	for _, line := range strings.Split(strings.ReplaceAll(string(payload), "\r\n", "\n"), "\n") {
		fmt.Fprintf(&buf, "data: %s\n", line)
	}

	if event.Retry > 0 {
		fmt.Fprintf(&buf, "retry: %d\n", event.Retry)
	}

	buf.WriteString("\n")

	_, err = io.WriteString(w, buf.String())
	return err
}
