package traffic

import "github.com/gin-gonic/gin"

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Status(c *gin.Context) {
	c.JSON(200, h.service.Update(0, 0, 0))
}

func (h *Handler) Users(c *gin.Context) {
	c.JSON(200, []UserTraffic{})
}

func (h *Handler) Connections(c *gin.Context) {
	c.JSON(200, gin.H{"connections": 0})
}
