package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers system stats routes under the provided group
func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/status", GetSystemStatus)
}

func GetSystemStatus(c *gin.Context) {
	stats := GlobalService.GetStatus()
	c.JSON(http.StatusOK, stats)
}
