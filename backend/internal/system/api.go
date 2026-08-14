package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes accepts an optional service. The GlobalService fallback is
// temporary compatibility for callers that have not yet migrated.
func RegisterRoutes(rg *gin.RouterGroup, services ...*Service) {
	svc := GlobalService
	if len(services) > 0 && services[0] != nil { svc = services[0] }
	rg.GET("/status", GetSystemStatus(svc))
}

func GetSystemStatus(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "system service not initialized"}); return }
		c.JSON(http.StatusOK, svc.GetStatus())
	}
}
