package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kazeyukiro/3m-ui/backend/internal/config"
	"github.com/kazeyukiro/3m-ui/backend/internal/converter"
	"github.com/kazeyukiro/3m-ui/backend/internal/database/models"
	"gorm.io/gorm"
)

// RegisterPublicSubscriptionRoutes mounts token-authenticated subscription
// endpoints before the administrator auth middleware. Possession of the
// high-entropy access token is the credential for these endpoints.
func RegisterPublicSubscriptionRoutes(api *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	api.GET("/client/sub/:token", func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database is not configured"})
			return
		}

		var token models.AccessToken
		if err := db.Where("token = ? AND enabled = ?", c.Param("token"), true).First(&token).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load subscription"})
			return
		}
		if token.ExpireAt != nil && !token.ExpireAt.After(time.Now()) {
			c.JSON(http.StatusGone, gin.H{"error": "subscription expired"})
			return
		}

		raw, err := converter.GenerateRawConfig(db, token, c.Request)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		// raw=true is used internally by subconverter as well as by callers
		// that explicitly request the canonical Mihomo YAML.
		target := strings.ToLower(strings.TrimSpace(c.Query("target")))
		if c.Query("raw") == "true" || target == "" || target == "mihomo" {
			c.Header("Cache-Control", "no-store")
			c.Data(http.StatusOK, "text/yaml; charset=utf-8", raw)
			return
		}

		// Other client formats are delegated to a local subconverter. The
		// canonical raw endpoint remains available even when subconverter is
		// not installed.
		converted, err := converter.CallSubconverterWithRequest(cfg, c.Param("token"), target, raw)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.Header("Cache-Control", "no-store")
		c.Data(http.StatusOK, "text/plain; charset=utf-8", converted)
	})
}
