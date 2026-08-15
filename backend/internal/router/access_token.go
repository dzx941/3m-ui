package router

import (
	"net/http"
	"strconv"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterAccessTokenRoutes(rg *gin.RouterGroup, d Deps) {
	db := resolveDB(d)
	h := &AccessTokenHandler{db: db}
	rg.GET("", h.List)
	rg.POST("", h.Create)
	rg.DELETE("/:id", h.Delete)
}

type AccessTokenHandler struct {
	db *gorm.DB
}

func (h *AccessTokenHandler) List(c *gin.Context) {
	var tokens []models.AccessToken
	if err := h.db.Find(&tokens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *AccessTokenHandler) Create(c *gin.Context) {
	var token models.AccessToken
	if err := c.ShouldBindJSON(&token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, token)
}

func (h *AccessTokenHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.db.Delete(&models.AccessToken{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func resolveDB(d Deps) *gorm.DB {
	if d.DB != nil {
		return d.DB
	}
	// Production environments must wire dependencies correctly.
	panic("database dependency is nil: router dependencies must be wired correctly")
}
