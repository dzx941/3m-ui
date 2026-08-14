package listener

import (
	"net/http"
	"strconv"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/gin-gonic/gin"
)

// Handler exposes listener HTTP endpoints with an explicitly injected
// listener service. Keeping the service on the handler removes the API
// layer's dependency on the package-level GlobalService.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers listener CRUD routes using the supplied service.
func RegisterRoutes(rg *gin.RouterGroup, service *Service) {
	h := NewHandler(service)
	rg.GET("", h.List)
	rg.POST("", h.Create)
	rg.GET("/:id", h.Get)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
	rg.POST("/:id/reload", h.Reload)
}

func (h *Handler) List(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "listener service not initialized"})
		return
	}
	list, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) Create(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "listener service not initialized"})
		return
	}
	var l models.Listener
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.Create(&l); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, l)
}

func (h *Handler) Get(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "listener service not initialized"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid listener id"})
		return
	}
	l, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, l)
}

func (h *Handler) Update(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "listener service not initialized"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid listener id"})
		return
	}
	var l models.Listener
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l.ID = uint(id)
	if err := h.service.Update(&l); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, l)
}

func (h *Handler) Delete(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "listener service not initialized"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid listener id"})
		return
	}
	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "listener deleted"})
}

func (h *Handler) Reload(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "listener service not initialized"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid listener id"})
		return
	}
	if err := h.service.TriggerReload(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "listener reloaded"})
}
