package node

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/converter"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", ListNodes)
	rg.POST("", CreateNode)
	rg.GET("/:id", GetNode)
	rg.PUT("/:id", UpdateNode)
	rg.DELETE("/:id", DeleteNode)
	rg.POST("/:id/reload", ReloadNode)
	rg.POST("/:id/client-access", CreateClientAccess)
}

func ListNodes(c *gin.Context) {
	list, err := GlobalService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func CreateNode(c *gin.Context) {
	var l models.Listener
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := GlobalService.Create(&l); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, l)
}

func GetNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"})
		return
	}

	l, err := GlobalService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, l)
}

func UpdateNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"})
		return
	}

	var l models.Listener
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	l.ID = uint(id)
	if err := GlobalService.Update(&l); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, l)
}

func DeleteNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"})
		return
	}

	if err := GlobalService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "node deleted"})
}

func ReloadNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"})
		return
	}

	if err := GlobalService.TriggerReload(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "node reloaded"})
}

func CreateClientAccess(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid listener id"})
		return
	}

	var listener models.Listener
	if err := database.GlobalDB.First(&listener, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listener not found"})
		return
	}
	if !listener.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "listener is disabled"})
		return
	}

	var existing models.AccessToken
	if err := database.GlobalDB.Where("scope = ? AND target_id = ? AND enabled = ?", "listener", listener.ID, true).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, clientAccessResponse(c, existing))
		return
	}

	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate secure token"})
		return
	}

	token := models.AccessToken{
		Name:     listener.Name,
		Token:    hex.EncodeToString(buf),
		Enabled:  true,
		Type:     "user", // Keep the public token API compatible; Scope distinguishes direct listener access.
		Scope:    "listener",
		TargetID: listener.ID,
	}
	if err := database.GlobalDB.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, clientAccessResponse(c, token))
}

func clientAccessResponse(c *gin.Context, token models.AccessToken) gin.H {
	return gin.H{
		"id":                token.ID,
		"name":              token.Name,
		"token":             token.Token,
		"type":               "listener",
		"scope":              token.Scope,
		"target_id":          token.TargetID,
		"mihomo_link":       converter.GetSubscriptionURL(config.GlobalConfig, c.Request, token.Token, "mihomo"),
		"clash_link":        converter.GetSubscriptionURL(config.GlobalConfig, c.Request, token.Token, "clash"),
		"singbox_link":      converter.GetSubscriptionURL(config.GlobalConfig, c.Request, token.Token, "singbox"),
		"shadowrocket_link": converter.GetSubscriptionURL(config.GlobalConfig, c.Request, token.Token, "shadowrocket"),
	}
}
