package router

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kazeyukiro/3m-ui/backend/internal/config"
	"github.com/kazeyukiro/3m-ui/backend/internal/converter"
	"github.com/kazeyukiro/3m-ui/backend/internal/database/models"
	"gorm.io/gorm"
)

func RegisterAccessTokenRoutes(rg *gin.RouterGroup, d Deps) {
	db := d.DB
	h := &AccessTokenHandler{db: db, cfg: d.Config}
	rg.GET("", h.List)
	rg.POST("", h.Create)
	rg.DELETE("/:id", h.Delete)
}

type AccessTokenHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

type accessTokenCreateInput struct {
	Name       string     `json:"name"`
	ListenerID uint       `json:"listener_id" binding:"required"`
	ExpireAt   *time.Time `json:"expire_at"`
}

type accessTokenResponse struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	Token      string     `json:"token"`
	Enabled    bool       `json:"enabled"`
	ExpireAt   *time.Time `json:"expire_at"`
	ListenerID uint       `json:"listener_id"`
}

func (h *AccessTokenHandler) List(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database is not configured"})
		return
	}
	var tokens []models.AccessToken
	if err := h.db.Order("id desc").Find(&tokens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]accessTokenResponse, 0, len(tokens))
	for _, token := range tokens {
		out = append(out, accessTokenResponse{
			ID: token.ID, Name: token.Name, Enabled: token.Enabled,
			ExpireAt: token.ExpireAt, ListenerID: token.ListenerID,
		})
	}
	c.JSON(http.StatusOK, out)
}

func (h *AccessTokenHandler) Create(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database is not configured"})
		return
	}
	var input accessTokenCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.ListenerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "listener_id is required"})
		return
	}
	if input.ExpireAt != nil && !input.ExpireAt.After(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expire_at must be in the future"})
		return
	}
	var listener models.Listener
	if err := h.db.First(&listener, input.ListenerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "listener not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !listener.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "listener is disabled"})
		return
	}

	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate secure token"})
		return
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = listener.Name
	}
	token := models.AccessToken{
		Name: name, Token: hex.EncodeToString(buf), Enabled: true,
		ExpireAt: input.ExpireAt, ListenerID: listener.ID,
	}
	if err := h.db.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": token.ID, "name": token.Name, "token": token.Token,
		"enabled": token.Enabled, "expire_at": token.ExpireAt,
		"listener_id":       token.ListenerID,
		"mihomo_link":       converter.GetSubscriptionURL(h.cfg, c.Request, token.Token, "mihomo"),
		"clash_link":        converter.GetSubscriptionURL(h.cfg, c.Request, token.Token, "clash"),
		"singbox_link":      converter.GetSubscriptionURL(h.cfg, c.Request, token.Token, "singbox"),
		"shadowrocket_link": converter.GetSubscriptionURL(h.cfg, c.Request, token.Token, "shadowrocket"),
	})
}

func (h *AccessTokenHandler) Delete(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database is not configured"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result := h.db.Delete(&models.AccessToken{}, uint(id))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "access token not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
