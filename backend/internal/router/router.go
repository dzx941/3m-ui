package router

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/auth"
	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/converter"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/listener"
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

func SetupRouter(cfg *config.Config, deps Dependencies) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

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
		auth.RegisterRoutes(apiV1.Group("/auth"), cfg)
		apiV1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		apiV1.GET("/client/sub/:token", func(c *gin.Context) {
			var token models.AccessToken
			if deps.DB == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
				return
			}
			if err := deps.DB.Where("token = ? AND enabled = ?", c.Param("token"), true).First(&token).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Token 不存在或已禁用"})
				return
			}
			if token.ExpireAt != nil && token.ExpireAt.Before(time.Now()) {
				c.JSON(http.StatusGone, gin.H{"error": "Token 已过期"})
				return
			}
			if token.ListenerID == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Token 未绑定 Listener"})
				return
			}

			rawYAML, err := converter.GenerateRawConfig(deps.DB, token, c.Request)
			if err != nil {
				status := http.StatusInternalServerError
				if err.Error() == "listener not found" || err.Error() == "listener is disabled" {
					status = http.StatusGone
				}
				c.JSON(status, gin.H{"error": err.Error()})
				return
			}

			c.Header("Content-Disposition", "attachment; filename=config.yaml")
			c.Data(http.StatusOK, "application/yaml; charset=utf-8", rawYAML)
		})

		apiV1.Use(auth.RequireAuth(cfg.JWT.Secret))

		accessTokenGroup := apiV1.Group("/access-tokens")
		{
			accessTokenGroup.GET("", func(c *gin.Context) {
				var tokens []models.AccessToken
				if deps.DB == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
					return
				}
				if err := deps.DB.Find(&tokens).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				type TokenResponse struct {
					models.AccessToken
					ListenerName     string `json:"listener_name"`
					ListenerProtocol string `json:"listener_protocol"`
					MihomoLink       string `json:"mihomo_link"`
					ClashLink        string `json:"clash_link"`
					SingboxLink      string `json:"singbox_link"`
					ShadowrocketLink string `json:"shadowrocket_link"`
				}

				resp := make([]TokenResponse, 0, len(tokens))
				for _, token := range tokens {
					item := TokenResponse{AccessToken: token}
					var l models.Listener
					if token.ListenerID != 0 && deps.DB.First(&l, token.ListenerID).Error == nil {
						item.ListenerName = l.Name
						item.ListenerProtocol = l.Protocol
					}
					link := converter.GetSubscriptionURL(cfg, c.Request, token.Token, "")
					item.MihomoLink = link
					item.ClashLink = link
					item.SingboxLink = link
					item.ShadowrocketLink = link
					resp = append(resp, item)
				}
				c.JSON(http.StatusOK, resp)
			})

			accessTokenGroup.POST("", func(c *gin.Context) {
				var req struct {
					Name       string     `json:"name" binding:"required"`
					ListenerID uint       `json:"listener_id" binding:"required"`
					ExpireAt   *time.Time `json:"expire_at"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				var l models.Listener
				if deps.DB == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
					return
				}
				if err := deps.DB.First(&l, req.ListenerID).Error; err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "listener not found"})
					return
				}
				if !l.Enabled {
					c.JSON(http.StatusBadRequest, gin.H{"error": "listener is disabled"})
					return
				}

				tokenObj := models.AccessToken{
					Name:       req.Name,
					Token:      generateSecureToken(),
					Enabled:    true,
					ExpireAt:   req.ExpireAt,
					ListenerID: req.ListenerID,
				}
				if err := deps.DB.Create(&tokenObj).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, tokenObj)
			})

			accessTokenGroup.PUT("/:id", func(c *gin.Context) {
				var req struct {
					Enabled bool `json:"enabled"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				var token models.AccessToken
				if deps.DB == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
					return
				}
				if err := deps.DB.First(&token, c.Param("id")).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
					return
				}
				token.Enabled = req.Enabled
				if err := deps.DB.Save(&token).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, token)
			})

			accessTokenGroup.DELETE("/:id", func(c *gin.Context) {
				var token models.AccessToken
				if deps.DB == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
					return
				}
				if err := deps.DB.First(&token, c.Param("id")).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
					return
				}
				if err := deps.DB.Delete(&token).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})
		}

		apiV1.GET("/dashboard", func(c *gin.Context) {
			var mihomoStatus *mihomo.StatusResponse
			var err error
			if deps.Mihomo != nil {
				mihomoStatus, err = deps.Mihomo.GetStatus()
			}
			if err != nil || mihomoStatus == nil {
				mihomoStatus = &mihomo.StatusResponse{Running: false, Version: "unknown", PID: 0, Uptime: "0s"}
			}
			sysStatus := system.GlobalService.GetStatus()

			var listenerTotal int64
			var listenerEnabled int64
			if deps.DB != nil {
				deps.DB.Model(&models.Listener{}).Count(&listenerTotal)
				deps.DB.Model(&models.Listener{}).Where("enabled = ?", true).Count(&listenerEnabled)
			}
			listenerDisabled := listenerTotal - listenerEnabled
			if listenerDisabled < 0 {
				listenerDisabled = 0
			}

			var trafficSnapshot traffic.Snapshot
			if deps.Traffic != nil {
				trafficSnapshot = deps.Traffic.Current()
			}
			var onlineUsers int64
			if deps.DB != nil {
				deps.DB.Model(&models.ProxyUser{}).Where("online = ?", true).Count(&onlineUsers)
			}
			activeConnections := trafficSnapshot.Connections
			if deps.Collector != nil {
				activeConnections = len(deps.Collector.CurrentConnections())
			}
			c.JSON(http.StatusOK, gin.H{
				"mihomo": mihomoStatus,
				"system": sysStatus,
				"listeners": gin.H{"total": listenerTotal, "enabled": listenerEnabled, "disabled": listenerDisabled},
				"traffic": gin.H{
					"uploadRate": trafficSnapshot.UploadRate,
					"downloadRate": trafficSnapshot.DownloadRate,
					"totalUpload": trafficSnapshot.UploadBytes,
					"totalDownload": trafficSnapshot.DownloadBytes,
					"onlineUsers": onlineUsers,
					"activeConnections": activeConnections,
				},
			})
		})

		systemGroup := apiV1.Group("/system")
		{
			system.RegisterRoutes(systemGroup)
		}

		mihomoGroup := apiV1.Group("/mihomo")
		{
			mihomoGroup.GET("/status", func(c *gin.Context) {
				if deps.Mihomo == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "mihomo service not initialized"}); return }
				status, err := deps.Mihomo.GetStatus()
				if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, status)
			})
			mihomoGroup.POST("/start", func(c *gin.Context) {
				if deps.Mihomo == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "mihomo service not initialized"}); return }
				if err := deps.Mihomo.StartMihomo(); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Mihomo started"})
			})
			mihomoGroup.POST("/stop", func(c *gin.Context) {
				if deps.Mihomo == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "mihomo service not initialized"}); return }
				if err := deps.Mihomo.StopMihomo(); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Mihomo stopped"})
			})
			mihomoGroup.POST("/restart", func(c *gin.Context) {
				if deps.Mihomo == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "mihomo service not initialized"}); return }
				if err := deps.Mihomo.RestartMihomo(); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Mihomo restarted"})
			})
			mihomoGroup.GET("/logs", func(c *gin.Context) {
				if deps.Mihomo == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "mihomo service not initialized"}); return }
				logs, err := deps.Mihomo.GetLogs()
				if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, logs)
			})
		}

		userGroup := apiV1.Group("/users")
		{
			user.RegisterRoutes(userGroup)
		}

		nodeGroup := apiV1.Group("/nodes")
		{
			node.RegisterRoutes(nodeGroup)
		}

		listenerGroup := apiV1.Group("/listeners")
		{
			listener.RegisterRoutes(listenerGroup, deps.Listener)
		}

		trafficGroup := apiV1.Group("/traffic")
		{
			trafficHandler := traffic.NewHandler(deps.Traffic, deps.Collector, deps.DB)
			traffic.RegisterRoutes(trafficGroup, trafficHandler)
		}

		configGroup := apiV1.Group("/config")
		{
			configGroup.GET("/proxies", func(c *gin.Context) {
				visual, err := mihomoConfig.GetVisualConfig(deps.DB)
				if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, visual.Proxies)
			})
			configGroup.POST("/proxies", func(c *gin.Context) {
				var proxy mihomoConfig.ProxyEntry
				if err := c.ShouldBindJSON(&proxy); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
				if proxy.Name == "" || proxy.Type == "" || proxy.Server == "" || proxy.Port < 1 || proxy.Port > 65535 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "名称、协议、服务器和有效端口不能为空"}); return
				}
				visual, err := mihomoConfig.GetVisualConfig(deps.DB)
				if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				visual.Proxies = append(visual.Proxies, proxy)
				if err = mihomoConfig.SaveVisualConfig(deps.DB, visual); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusCreated, proxy)
			})
			configGroup.PUT("/proxies/:index", func(c *gin.Context) {
				idx, err := strconv.Atoi(c.Param("index"))
				if err != nil || idx < 0 { c.JSON(http.StatusBadRequest, gin.H{"error": "无效节点索引"}); return }
				var proxy mihomoConfig.ProxyEntry
				if err = c.ShouldBindJSON(&proxy); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
				visual, err := mihomoConfig.GetVisualConfig(deps.DB)
				if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				if idx >= len(visual.Proxies) { c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"}); return }
				visual.Proxies[idx] = proxy
				if err = mihomoConfig.SaveVisualConfig(deps.DB, visual); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, proxy)
			})
			configGroup.DELETE("/proxies/:index", func(c *gin.Context) {
				idx, err := strconv.Atoi(c.Param("index"))
				if err != nil || idx < 0 { c.JSON(http.StatusBadRequest, gin.H{"error": "无效节点索引"}); return }
				visual, err := mihomoConfig.GetVisualConfig(deps.DB)
				if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				if idx >= len(visual.Proxies) { c.JSON(http.StatusNotFound, gin.H{"error": "节点不存在"}); return }
				visual.Proxies = append(visual.Proxies[:idx], visual.Proxies[idx+1:]...)
				if err = mihomoConfig.SaveVisualConfig(deps.DB, visual); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})
			configGroup.GET("/visual", func(c *gin.Context) {
				visual, err := mihomoConfig.GetVisualConfig(deps.DB)
				if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, visual)
			})
			configGroup.POST("/visual", func(c *gin.Context) {
				var visual mihomoConfig.VisualConfig
				if err := c.ShouldBindJSON(&visual); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visual configuration: " + err.Error()}); return }
				if err := mihomoConfig.SaveVisualConfig(deps.DB, visual); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Visual configuration saved"})
			})
			configGroup.GET("", func(c *gin.Context) {
				engine := deps.ConfigEngine
				if engine == nil { engine = mihomoConfig.NewConfigEngine(deps.DB) }
				yamlStr, err := engine.GenerateFinalConfig()
				if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, gin.H{"config": yamlStr})
			})
			configGroup.POST("/generate", func(c *gin.Context) {
				engine := deps.ConfigEngine
				if engine == nil { engine = mihomoConfig.NewConfigEngine(deps.DB) }
				yamlStr, err := engine.GenerateFinalConfig()
				if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				dir := filepath.Dir(cfg.Mihomo.Config)
				if err := os.MkdirAll(dir, 0755); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				if err := os.WriteFile(cfg.Mihomo.Config, []byte(yamlStr), 0644); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Configuration generated and written successfully"})
			})
			configGroup.POST("/validate", func(c *gin.Context) {
				var req struct { Config string `json:"config"` }
				if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"}); return }
				if err := mihomoConfig.ValidateConfigYAML(req.Config); err != nil { c.JSON(http.StatusOK, gin.H{"valid": false, "error": err.Error()}); return }
				c.JSON(http.StatusOK, gin.H{"valid": true})
			})
			configGroup.GET("/download", func(c *gin.Context) {
				engine := deps.ConfigEngine
				if engine == nil { engine = mihomoConfig.NewConfigEngine(deps.DB) }
				yamlStr, err := engine.GenerateFinalConfig()
				if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
				c.Header("Content-Disposition", "attachment; filename=config.yaml")
				c.Data(http.StatusOK, "application/yaml", []byte(yamlStr))
			})
		}
	}

	return r
}
