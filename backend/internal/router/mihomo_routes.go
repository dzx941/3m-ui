package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func registerMihomoRoutes(api *gin.RouterGroup, d Deps) {
	group := api.Group("/mihomo")
	group.GET("/status", func(c *gin.Context) {
		status, err := d.mihomoService().GetStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, status)
	})
	group.POST("/start", func(c *gin.Context) {
		if err := d.mihomoService().StartMihomo(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Mihomo started"})
	})
	group.POST("/stop", func(c *gin.Context) {
		if err := d.mihomoService().StopMihomo(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Mihomo stopped"})
	})
	group.POST("/restart", func(c *gin.Context) {
		if err := d.mihomoService().RestartMihomo(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Mihomo restarted"})
	})
	group.GET("/logs", func(c *gin.Context) {
		logs, err := d.mihomoService().GetLogs()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, logs)
	})
}
