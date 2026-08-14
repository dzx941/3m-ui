package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(rg *gin.RouterGroup, svc *Service) {
	rg.GET("", List(svc))
	rg.POST("", Create(svc))
	rg.GET("/:id", Get(svc))
	rg.PUT("/:id", Update(svc))
	rg.DELETE("/:id", Delete(svc))
	rg.POST("/:id/listeners", BindListeners(svc))
	rg.GET("/:id/listeners", GetListeners(svc))
	// /nodes is the public naming used by the server-node UI; /listeners is
	// retained as a backward-compatible alias.
	rg.POST("/:id/nodes", BindListeners(svc))
	rg.GET("/:id/nodes", GetListeners(svc))
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return 0, false
	}
	return uint(id), true
}

func List(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "user service not initialized"}); return }
		users, err := svc.GetAll()
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
		out := make([]SafeUser, 0, len(users))
		for i := range users { out = append(out, ToSafeUser(&users[i])) }
		c.JSON(http.StatusOK, out)
	}
}

func Create(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "user service not initialized"}); return }
		var in CreateInput
		if err := c.ShouldBindJSON(&in); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
		u, err := svc.Create(in)
		if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusCreated, ToSafeUser(u))
	}
}

func Get(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "user service not initialized"}); return }
		id, ok := parseID(c); if !ok { return }
		u, err := svc.GetByID(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) { c.JSON(http.StatusNotFound, gin.H{"error": "user not found"}) } else { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) }
			return
		}
		c.JSON(http.StatusOK, ToSafeUser(u))
	}
}

func Update(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "user service not initialized"}); return }
		id, ok := parseID(c); if !ok { return }
		var in UpdateInput
		if err := c.ShouldBindJSON(&in); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
		u, err := svc.Update(id, in)
		if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusOK, ToSafeUser(u))
	}
}

func Delete(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "user service not initialized"}); return }
		id, ok := parseID(c); if !ok { return }
		if err := svc.Delete(id); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func BindListeners(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "user service not initialized"}); return }
		id, ok := parseID(c); if !ok { return }
		var req struct { ListenerIDs []uint `json:"listener_ids"` }
		if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
		if err := svc.BindListeners(id, req.ListenerIDs); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
		c.JSON(http.StatusOK, gin.H{"status": "ok", "listener_ids": req.ListenerIDs})
	}
}

func GetListeners(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "user service not initialized"}); return }
		id, ok := parseID(c); if !ok { return }
		list, err := svc.GetListeners(id)
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
		type nodeDTO struct { ID uint `json:"id"`; Name string `json:"name"`; Protocol string `json:"protocol"`; Port int `json:"port"`; BindAddress string `json:"bind_address"`; Enabled bool `json:"enabled"`; TLS bool `json:"tls"`; UDP bool `json:"udp"`; Status string `json:"status"` }
		out := make([]nodeDTO, 0, len(list))
		for _, n := range list { out = append(out, nodeDTO{n.ID, n.Name, n.Protocol, n.Port, n.BindAddress, n.Enabled, n.TLS, n.UDP, n.Status}) }
		c.JSON(http.StatusOK, out)
	}
}

// Keep models import referenced for future DTO extensions and schema compatibility.
var _ = models.Listener{}
