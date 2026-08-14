package node

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/converter"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(rg *gin.RouterGroup, svc *Service, db *gorm.DB, cfg *config.Config, userSvc *user.Service) {
	rg.GET("", ListNodes(svc))
	rg.POST("", CreateNode(svc))
	rg.GET("/:id", GetNode(svc))
	rg.GET("/:id/uri", ExportNodeURI(svc, userSvc))
	rg.PUT("/:id", UpdateNode(svc))
	rg.DELETE("/:id", DeleteNode(svc))
	rg.POST("/:id/reload", ReloadNode(svc))
	rg.POST("/:id/client-access", CreateClientAccess(db, cfg))
}

func ListNodes(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "node service not initialized"}); return }
		list, err := svc.GetAll()
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusOK, list)
	}
}

func CreateNode(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "node service not initialized"}); return }
		var l models.Listener
		if err := c.ShouldBindJSON(&l); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
		if err := ValidateNode(&l); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
		if err := svc.Create(&l); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusCreated, l)
	}
}

func GetNode(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "node service not initialized"}); return }
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"}); return }
		l, err := svc.GetByID(uint(id))
		if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusOK, l)
	}
}

// ExportNodeURI returns one URI per active client credential. Public host
// comes from the request Host; wildcard listener addresses are never exposed.
func ExportNodeURI(svc *Service, userSvc *user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "node service not initialized"}); return }
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"}); return }
		listener, err := svc.GetByID(uint(id))
		if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": "listener not found"}); return }
		if !listener.Enabled { c.JSON(http.StatusBadRequest, gin.H{"error": "listener is disabled"}); return }

		host := c.GetHeader("X-Forwarded-Host")
		if host == "" { host = c.Request.Host }
		if host != "" {
			if parsed, err := url.Parse("//" + host); err == nil && parsed.Hostname() != "" { host = parsed.Hostname() }
		}

		credentials := []user.Credential{}
		if userSvc != nil {
			byListener, err := userSvc.ActiveCredentialsByListener()
			if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load active credentials: " + err.Error()}); return }
			credentials = byListener[listener.ID]
		}

		uris, err := ClientURIsWithCredentials(*listener, strings.TrimSpace(host), credentials)
		if err != nil { c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusOK, gin.H{"name": listener.Name, "protocol": listener.Protocol, "uris": uris})
	}
}

func UpdateNode(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "node service not initialized"}); return }
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"}); return }
		var l models.Listener
		if err := c.ShouldBindJSON(&l); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
		l.ID = uint(id)
		if err := ValidateNode(&l); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
		if err := svc.Update(&l); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusOK, l)
	}
}

func DeleteNode(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "node service not initialized"}); return }
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"}); return }
		if err := svc.Delete(uint(id)); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "node deleted"})
	}
}

func ReloadNode(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "node service not initialized"}); return }
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"}); return }
		if err := svc.TriggerReload(uint(id)); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "node reloaded"})
	}
}

func CreateClientAccess(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil || cfg == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "node dependencies not initialized"}); return }
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid listener id"}); return }
		var listener models.Listener
		if err := db.First(&listener, uint(id)).Error; err != nil { c.JSON(http.StatusNotFound, gin.H{"error": "listener not found"}); return }
		if !listener.Enabled { c.JSON(http.StatusBadRequest, gin.H{"error": "listener is disabled"}); return }
		var existing models.AccessToken
		if err := db.Where("listener_id = ? AND enabled = ?", listener.ID, true).First(&existing).Error; err == nil {
			if existing.ExpireAt == nil || existing.ExpireAt.After(time.Now()) { c.JSON(http.StatusOK, clientAccessResponse(c, existing, cfg)); return }
		}
		buf := make([]byte, 16)
		if _, err := rand.Read(buf); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate secure token"}); return }
		token := models.AccessToken{Name: listener.Name, Token: hex.EncodeToString(buf), Enabled: true, ListenerID: listener.ID}
		if err := db.Create(&token).Error; err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusCreated, clientAccessResponse(c, token, cfg))
	}
}

func clientAccessResponse(c *gin.Context, token models.AccessToken, cfg *config.Config) gin.H {
	return gin.H{
		"id": token.ID, "name": token.Name, "token": token.Token, "type": "listener", "listener_id": token.ListenerID,
		"mihomo_link": converter.GetSubscriptionURL(cfg, c.Request, token.Token, "mihomo"),
		"clash_link": converter.GetSubscriptionURL(cfg, c.Request, token.Token, "clash"),
		"singbox_link": converter.GetSubscriptionURL(cfg, c.Request, token.Token, "singbox"),
		"shadowrocket_link": converter.GetSubscriptionURL(cfg, c.Request, token.Token, "shadowrocket"),
	}
}

// Keep the database package referenced by this file's public compatibility
// surface during the service migration.
var _ = database.GlobalDB
