package router

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/listener"
	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// CORS Middleware (simple for development)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})

		// Mihomo Core Management APIs
		mihomoGroup := apiV1.Group("/mihomo")
		{
			mihomoGroup.GET("/status", func(c *gin.Context) {
				status, err := mihomo.GlobalService.GetStatus()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, status)
			})

			mihomoGroup.POST("/start", func(c *gin.Context) {
				err := mihomo.GlobalService.StartMihomo()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"status":  "ok",
					"message": "Mihomo started",
				})
			})

			mihomoGroup.POST("/stop", func(c *gin.Context) {
				err := mihomo.GlobalService.StopMihomo()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"status":  "ok",
					"message": "Mihomo stopped",
				})
			})

			mihomoGroup.POST("/restart", func(c *gin.Context) {
				err := mihomo.GlobalService.RestartMihomo()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"status":  "ok",
					"message": "Mihomo restarted",
				})
			})

			mihomoGroup.GET("/logs", func(c *gin.Context) {
				logs, err := mihomo.GlobalService.GetLogs()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, logs)
			})
		}

		// Mihomo Inbound/Listener APIs
		listenerGroup := apiV1.Group("/listeners")
		{
			listener.RegisterRoutes(listenerGroup)
		}

		// Config Engine APIs
		configGroup := apiV1.Group("/config")
		{
			configGroup.GET("", func(c *gin.Context) {
				engine := mihomoConfig.NewConfigEngine(database.GlobalDB)
				yamlStr, err := engine.GenerateFinalConfig()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"config": yamlStr,
				})
			})

			configGroup.POST("/generate", func(c *gin.Context) {
				engine := mihomoConfig.NewConfigEngine(database.GlobalDB)
				yamlStr, err := engine.GenerateFinalConfig()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				// Write config to designated path
				dir := filepath.Dir(cfg.Mihomo.Config)
				if err := os.MkdirAll(dir, 0755); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				if err := os.WriteFile(cfg.Mihomo.Config, []byte(yamlStr), 0644); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status":  "ok",
					"message": "Configuration generated and written successfully",
				})
			})

			configGroup.POST("/validate", func(c *gin.Context) {
				var req struct {
					Config string `json:"config"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
					return
				}

				err := mihomoConfig.ValidateConfigYAML(req.Config)
				if err != nil {
					c.JSON(http.StatusOK, gin.H{
						"valid": false,
						"error": err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"valid": true,
				})
			})

			configGroup.GET("/download", func(c *gin.Context) {
				// Generate dynamically or serve existing file. Best practice: generate latest dynamically!
				engine := mihomoConfig.NewConfigEngine(database.GlobalDB)
				yamlStr, err := engine.GenerateFinalConfig()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.Header("Content-Disposition", "attachment; filename=config.yaml")
				c.Data(http.StatusOK, "application/yaml", []byte(yamlStr))
			})
		}
	}

	return r
}
