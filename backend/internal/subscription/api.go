package subscription

import (
	"net/http"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/sub/:token", func(c *gin.Context) {
		var sub models.Subscription
		if err := database.GlobalDB.Where("token = ?", c.Param("token")).First(&sub).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}

		var data []byte
		gen := NewGenerator(database.GlobalDB)
		data, err := gen.Generate(sub.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if c.Query("format") == "base64" {
			c.String(http.StatusOK, EncodeBase64(data))
			return
		}
		c.Data(http.StatusOK, "text/yaml; charset=utf-8", data)
	})

	r.POST("/subscriptions", func(c *gin.Context) {
		var req struct {
			UserID uint `json:"user_id"`
			Format string `json:"format"`
		}
		if c.ShouldBindJSON(&req) != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":"invalid request"})
			return
		}
		s := NewService(database.GlobalDB)
		sub, err := s.Create(req.UserID, strings.ToLower(req.Format))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		c.JSON(http.StatusOK, sub)
	})
}
