package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers system stats routes under the provided group.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service) {
	rg.GET("/status", GetSystemStatus(svc))
}

func GetSystemStatus(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "system service not initialized"}); return }
		stats := svc.GetStatus()
		c.JSON(http.StatusOK, stats)
	}
}
