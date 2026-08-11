package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", List)
	rg.POST("", Create)
	rg.GET("/:id", Get)
	rg.PUT("/:id", Update)
	rg.DELETE("/:id", Delete)
	rg.POST("/:id/listeners", BindListeners)
	rg.GET("/:id/listeners", GetListeners)
	// /nodes is the public naming used by the server-node UI; /listeners is
	// retained as a backward-compatible alias.
	rg.POST("/:id/nodes", BindListeners)
	rg.GET("/:id/nodes", GetListeners)
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return 0, false
	}
	return uint(id), true
}

func List(c *gin.Context) {
	users, err := GlobalService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]SafeUser, 0, len(users))
	for i := range users {
		out = append(out, ToSafeUser(&users[i]))
	}
	c.JSON(http.StatusOK, out)
}

func Create(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := GlobalService.Create(in)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ToSafeUser(u))
}

func Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	u, err := GlobalService.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, ToSafeUser(u))
}

func Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := GlobalService.Update(id, in)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ToSafeUser(u))
}

func Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := GlobalService.Delete(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func BindListeners(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req struct {
		ListenerIDs []uint `json:"listener_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := GlobalService.BindListeners(id, req.ListenerIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "listener_ids": req.ListenerIDs})
}

func GetListeners(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	list, err := GlobalService.GetListeners(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Return only non-sensitive node fields.
	type nodeDTO struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Protocol    string `json:"protocol"`
		Port        int    `json:"port"`
		BindAddress string `json:"bind_address"`
		Enabled     bool   `json:"enabled"`
		TLS         bool   `json:"tls"`
		UDP         bool   `json:"udp"`
		Status      string `json:"status"`
	}
	out := make([]nodeDTO, 0, len(list))
	for _, n := range list {
		out = append(out, nodeDTO{n.ID, n.Name, n.Protocol, n.Port, n.BindAddress, n.Enabled, n.TLS, n.UDP, n.Status})
	}
	c.JSON(http.StatusOK, out)
}

// Keep models import referenced for future DTO extensions and schema compatibility.
var _ = models.Listener{}
