package router

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/dzx941/3m-ui/backend/internal/auth"
	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
	"github.com/dzx941/3m-ui/backend/internal/node"
	"github.com/dzx941/3m-ui/backend/internal/subscription"
	"github.com/dzx941/3m-ui/backend/internal/system"
	"github.com/dzx941/3m-ui/backend/internal/traffic"
	"github.com/dzx941/3m-ui/backend/internal/user"
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
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
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
		// Authentication endpoints are public; all management APIs below are protected.
		auth.RegisterRoutes(apiV1.Group("/auth"), cfg)
		apiV1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})

		// Everything after health is management API and requires a valid JWT.
		apiV1.Use(auth.RequireAuth(cfg.JWT.Secret))
		subscription.RegisterRoutes(apiV1)

		// Unified Dashboard Aggregator Endpoint
		apiV1.GET("/dashboard", func(c *gin.Context) {
			// 1. Mihomo core status
			mihomoStatus, err := mihomo.GlobalService.GetStatus()
			if err != nil {
				mihomoStatus = &mihomo.StatusResponse{Running: false, Version: "unknown", PID: 0, Uptime: "0s"}
			}

			// 2. System performance metrics
			sysStatus := system.GlobalService.GetStatus()

			// 3. Listener (Server Node) Statistics
			var totalCount, enabledCount, disabledCount int64
			if database.GlobalDB != nil {
				database.GlobalDB.Model(&models.Listener{}).Count(&totalCount)
				database.GlobalDB.Model(&models.Listener{}).Where("enabled = ?", true).Count(&enabledCount)
				database.GlobalDB.Model(&models.Listener{}).Where("enabled = ?", false).Count(&disabledCount)
			}

			// 4. Traffic statistics (Phase 8.5). Guarded against the traffic
			// scheduler not being initialized (e.g. in tests that construct
			// the router directly) so this never panics.
			var trafficSnapshot traffic.Snapshot
			if traffic.GlobalService != nil {
				trafficSnapshot = traffic.GlobalService.Current()
			}
			var onlineUsers int64
			if database.GlobalDB != nil {
				database.GlobalDB.Model(&models.ProxyUser{}).Where("online = ?", true).Count(&onlineUsers)
			}
			activeConnections := trafficSnapshot.Connections
			if traffic.GlobalCollector != nil {
				activeConnections = len(traffic.GlobalCollector.CurrentConnections())
			}

			c.JSON(http.StatusOK, gin.H{
				"mihomo": mihomoStatus,
				"system": sysStatus,
				"listeners": gin.H{
					"total":    totalCount,
					"enabled":  enabledCount,
					"disabled": disabledCount,
				},
				"traffic": gin.H{
					"uploadRate":        trafficSnapshot.UploadRate,
					"downloadRate":      trafficSnapshot.DownloadRate,
					"totalUpload":       trafficSnapshot.UploadBytes,
					"totalDownload":     trafficSnapshot.DownloadBytes,
					"onlineUsers":       onlineUsers,
					"activeConnections": activeConnections,
				},
			})
		})

		// System Performance APIs
		systemGroup := apiV1.Group("/system")
		{
			system.RegisterRoutes(systemGroup)
		}

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

		// Proxy User management routes. These users authenticate to Mihomo nodes;
		// they are intentionally separate from the 3m-ui administrator User model.
		userGroup := apiV1.Group("/users")
		{
			user.RegisterRoutes(userGroup)
		}

		// Server Node management routes
		nodeGroup := apiV1.Group("/nodes")
		{
			node.RegisterRoutes(nodeGroup)
		}

		// Mihomo Inbound/Listener APIs (backward-compatible, points to node package routes)
		listenerGroup := apiV1.Group("/listeners")
		{
			node.RegisterRoutes(listenerGroup)
		}

		// Traffic monitoring APIs (Phase 8.5)
		trafficGroup := apiV1.Group("/traffic")
		{
			trafficHandler := traffic.NewHandler(traffic.GlobalService, traffic.GlobalCollector, database.GlobalDB)
			traffic.RegisterRoutes(trafficGroup, trafficHandler)
		}

		// Config Engine APIs
		configGroup := apiV1.Group("/config")
		{
			// Visual configuration endpoints used by the React configuration UI.
			configGroup.GET("/visual", func(c *gin.Context) {
				visual, err := mihomoConfig.GetVisualConfig(database.GlobalDB)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, visual)
			})

			configGroup.POST("/visual", func(c *gin.Context) {
				var visual mihomoConfig.VisualConfig
				if err := c.ShouldBindJSON(&visual); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visual configuration: " + err.Error()})
					return
				}
				if err := mihomoConfig.SaveVisualConfig(database.GlobalDB, visual); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Visual configuration saved"})
			})

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
