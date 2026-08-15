package router

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/converter"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func generateSecureToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func resolveDB(d Deps) *gorm.DB {
	if d.DB != nil {
		return d.DB
	}
	return database.GlobalDB
}

func registerPublicClientRoutes(api *gin.RouterGroup, d Deps) {
	api.GET("/client/sub/:token", func(c *gin.Context) {
		db := resolveDB(d)
		var token models.AccessToken
		if err := db.Where("token = ? AND enabled = ?", c.Param("token"), true).First(&token).Error; err != nil {
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

		rawYAML, err := converter.GenerateRawConfig(db, token, c.Request)
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
}

func registerAccessTokenRoutes(api *gin.RouterGroup, d Deps, cfg *config.Config) {
	db := resolveDB(d)
	group := api.Group("/access-tokens")

	group.GET("", func(c *gin.Context) {
		var tokens []models.AccessToken
		if err := db.Find(&tokens).Error; err != nil {
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
			var listener models.Listener
			if token.ListenerID != 0 && db.First(&listener, token.ListenerID).Error == nil {
				item.ListenerName = listener.Name
				item.ListenerProtocol = listener.Protocol
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

	group.POST("", func(c *gin.Context) {
		var req struct {
			Name       string     `json:"name" binding:"required"`
			ListenerID uint       `json:"listener_id" binding:"required"`
			ExpireAt   *time.Time `json:"expire_at"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var listener models.Listener
		if err := db.First(&listener, req.ListenerID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "listener not found"})
			return
		}
		if !listener.Enabled {
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
		if err := db.Create(&tokenObj).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, tokenObj)
	})

	group.PUT("/:id", func(c *gin.Context) {
		var req struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var token models.AccessToken
		if err := db.First(&token, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			return
		}
		token.Enabled = req.Enabled
		if err := db.Save(&token).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, token)
	})

	group.DELETE("/:id", func(c *gin.Context) {
		var token models.AccessToken
		if err := db.First(&token, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			return
		}
		if err := db.Delete(&token).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
