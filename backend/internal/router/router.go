package router

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/auth"
	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/converter"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
	"github.com/dzx941/3m-ui/backend/internal/node"
	"github.com/dzx941/3m-ui/backend/internal/system"
	"github.com/dzx941/3m-ui/backend/internal/traffic"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"github.com/gin-gonic/gin"
)

func generateSecureToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

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

		// Public client subscription route (Restructured Subscription System)
		apiV1.GET("/client/sub/:token", func(c *gin.Context) {
			// Resolve token
			var token models.AccessToken
			if err := database.GlobalDB.Where("token = ? AND enabled = ?", c.Param("token"), true).First(&token).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Token 不存在或已禁用"})
				return
			}

			// Check expiration
			if token.ExpireAt != nil && token.ExpireAt.Before(time.Now()) {
				c.JSON(http.StatusGone, gin.H{"error": "Token 已过期"})
				return
			}

			// Validate Type
			if token.Type != "user" && token.Type != "proxy" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Token 类型无效"})
				return
			}

			// Generate Clash/Mihomo raw configuration for this token
			rawYAML, err := converter.GenerateRawConfig(database.GlobalDB, token, c.Request)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			target := c.Query("target")
			rawOnly := c.Query("raw") == "true"

			// If target is empty, or target is clash/mihomo, or raw is true, return raw YAML directly
			if rawOnly || target == "" || target == "clash" || target == "mihomo" {
				c.Header("Content-Disposition", "attachment; filename=config.yaml")
				c.Data(http.StatusOK, "application/yaml; charset=utf-8", rawYAML)
				return
			}

			// Call subconverter
			converted, err := converter.CallSubconverter(cfg, token.Token, target, rawYAML)
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "subconverter 不可用: " + err.Error()})
				return
			}

			// Return the converted config
			contentType := "text/plain; charset=utf-8"
			filename := "config.txt"
			if target == "singbox" {
				contentType = "application/json; charset=utf-8"
				filename = "config.json"
			}
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
			c.Data(http.StatusOK, contentType, converted)
		})

		// Everything after health and sub is management API and requires a valid JWT.
		apiV1.Use(auth.RequireAuth(cfg.JWT.Secret))

		// Access Token management routes (Phase 2 Restructuring)
		accessTokenGroup := apiV1.Group("/access-tokens")
		{
			// List all access tokens
			accessTokenGroup.GET("", func(c *gin.Context) {
				var tokens []models.AccessToken
				if err := database.GlobalDB.Find(&tokens).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				// Enriched response with dynamic sub URLs for UI use
				type TokenResponse struct {
					models.AccessToken
					MihomoLink       string `json:"mihomo_link"`
					ClashLink        string `json:"clash_link"`
					SingboxLink      string `json:"singbox_link"`
					ShadowrocketLink string `json:"shadowrocket_link"`
				}

				resp := make([]TokenResponse, 0, len(tokens))
				for _, t := range tokens {
					resp = append(resp, TokenResponse{
						AccessToken:      t,
						MihomoLink:       converter.GetSubscriptionURL(cfg, c.Request, t.Token, ""),
						ClashLink:        converter.GetSubscriptionURL(cfg, c.Request, t.Token, "clash"),
						SingboxLink:      converter.GetSubscriptionURL(cfg, c.Request, t.Token, "singbox"),
						ShadowrocketLink: converter.GetSubscriptionURL(cfg, c.Request, t.Token, "shadowrocket"),
					})
				}

				c.JSON(http.StatusOK, resp)
			})

			// Create a new access token
			accessTokenGroup.POST("", func(c *gin.Context) {
				var req struct {
					Name     string     `json:"name" binding:"required"`
					Type     string     `json:"type" binding:"required"`
					TargetID uint       `json:"target_id" binding:"required"`
					ExpireAt *time.Time `json:"expire_at"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if req.Type != "user" && req.Type != "proxy" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "type must be either 'user' or 'proxy'"})
					return
				}

				// Validate target existence
				if req.Type == "user" {
					var u models.ProxyUser
					if err := database.GlobalDB.First(&u, req.TargetID).Error; err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": "proxy user not found"})
						return
					}
				} else {
					// Validate Proxy Node index exists
					visual, err := mihomoConfig.GetVisualConfig(database.GlobalDB)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load visual config"})
						return
					}
					if int(req.TargetID) >= len(visual.Proxies) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "proxy node index out of range"})
						return
					}
				}

				token := generateSecureToken()

				tokenObj := models.AccessToken{
					Name:     req.Name,
					Token:    token,
					Enabled:  true,
					ExpireAt: req.ExpireAt,
					Type:     req.Type,
					TargetID: req.TargetID,
				}

				if err := database.GlobalDB.Create(&tokenObj).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, tokenObj)
			})

			// Toggle enabled/disabled
			accessTokenGroup.PUT("/:id", func(c *gin.Context) {
				var req struct {
					Enabled bool `json:"enabled"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				id := c.Param("id")
				var token models.AccessToken
				if err := database.GlobalDB.First(&token, id).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
					return
				}
				token.Enabled = req.Enabled
				if err := database.GlobalDB.Save(&token).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, token)
			})

			// Delete an access token
			accessTokenGroup.DELETE("/:id", func(c *gin.Context) {
				id := c.Param("id")
				if err := database.GlobalDB.Delete(&models.AccessToken{}, id).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})
		}

		// Unified Dashboard Aggregator Endpoint
		apiV1.GET("/dashboard", func(c *gin.Context) {
			// 1. Mihomo core status
			mihomoStatus, err := mihomo.GlobalService.GetStatus()
			if err != nil {
				mihomoStatus = &mihomo.StatusResponse{Running: false, Version: "unknown", PID: 0, Uptime: "0s"}
			}

			// 2. System performance metrics
			sysStatus := system.GlobalService.GetStatus()

			// 3. Proxy statistics come from the visual Mihomo configuration.
			visualCfg, _ := mihomoConfig.GetVisualConfig(database.GlobalDB)
			proxyCount := int64(len(visualCfg.Proxies))

			// 4. Traffic statistics. Guarded against the traffic
			// scheduler not being initialized so this never panics.
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
					"total":     proxyCount,
					"enabled":   proxyCount,
					"disabled":  0,
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

		// Proxy User management routes.
		userGroup := apiV1.Group("/users")
		{
			user.RegisterRoutes(userGroup)
		}

		// Server Node management routes
		nodeGroup := apiV1.Group("/nodes")
		{
			node.RegisterRoutes(nodeGroup)
		}

		// Mihomo Inbound/Listener APIs
		listenerGroup := apiV1.Group("/listeners")
		{
			node.RegisterRoutes(listenerGroup)
		}

		// Traffic monitoring APIs
		trafficGroup := apiV1.Group("/traffic")
		{
			trafficHandler := traffic.NewHandler(traffic.GlobalService, traffic.GlobalCollector, database.GlobalDB)
			traffic.RegisterRoutes(trafficGroup, trafficHandler)
		}

		// Config Engine APIs
		configGroup := apiV1.Group("/config")
		{
			configGroup.GET("/proxies", func(c *gin.Context) {
				visual, err := mihomoConfig.GetVisualConfig(database.GlobalDB)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, visual.Proxies)
			})
			configGroup.POST("/proxies", func(c *gin.Context) {
				var proxy mihomoConfig.ProxyEntry
				if err := c.ShouldBindJSON(&proxy); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				if proxy.Name == "" || proxy.Type == "" || proxy.Server == "" || proxy.Port < 1 || proxy.Port > 65535 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "名称、协议、服务器和有效端口不能为空"})
					return
				}
				visual, err := mihomoConfig.GetVisualConfig(database.GlobalDB)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				visual.Proxies = append(visual.Proxies, proxy)
				if err = mihomoConfig.SaveVisualConfig(database.GlobalDB, visual); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, proxy)
			})
			configGroup.PUT("/proxies/:index", func(c *gin.Context) {
				idx, err := strconv.Atoi(c.Param("index"))
				if err != nil || idx < 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效节点索引"})
					return
				}
				var proxy mihomoConfig.ProxyEntry
				if err = c.ShouldBindJSON(&proxy); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				visual, err := mihomoConfig.GetVisualConfig(database.GlobalDB)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				if idx >= len(visual.Proxies) {
					c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
					return
				}
				visual.Proxies[idx] = proxy
				if err = mihomoConfig.SaveVisualConfig(database.GlobalDB, visual); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, proxy)
			})
			configGroup.DELETE("/proxies/:index", func(c *gin.Context) {
				idx, err := strconv.Atoi(c.Param("index"))
				if err != nil || idx < 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效节点索引"})
					return
				}
				visual, err := mihomoConfig.GetVisualConfig(database.GlobalDB)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				if idx >= len(visual.Proxies) {
					c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"})
					return
				}
				visual.Proxies = append(visual.Proxies[:idx], visual.Proxies[idx+1:]...)
				if err = mihomoConfig.SaveVisualConfig(database.GlobalDB, visual); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

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
