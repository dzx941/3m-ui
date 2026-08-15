package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler serves system HTTP endpoints using an injected Service.
type Handler struct {
	svc *Service
}

// NewHandler constructs a system HTTP handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers system stats routes under the provided group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/status", h.GetSystemStatus)
}

func (h *Handler) GetSystemStatus(c *gin.Context) {
	stats := h.svc.GetStatus()
	c.JSON(http.StatusOK, stats)
}
