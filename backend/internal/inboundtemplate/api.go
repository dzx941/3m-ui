package inboundtemplate

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kazeyukiro/3m-ui/backend/internal/database/models"
	"github.com/kazeyukiro/3m-ui/backend/internal/user"
)

type Handler struct {
	nodeCreate func(*models.Listener) error
	userCreate func(user.CreateInput) (*models.ProxyUser, error)
	bind       func(uint, []uint) error
}

func NewHandler(nodeCreate func(*models.Listener) error, userCreate func(user.CreateInput) (*models.ProxyUser, error), bind func(uint, []uint) error) *Handler {
	return &Handler{nodeCreate: nodeCreate, userCreate: userCreate, bind: bind}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.List)
	rg.POST("/:id/create", h.Create)
}

func (h *Handler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"templates": List()})
}

func (h *Handler) Create(c *gin.Context) {
	var input CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.TemplateID == "" { input.TemplateID = c.Param("id") }
	listener, proxyUser, err := Create(input, h.nodeCreate, h.userCreate, h.bind)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"listener": listener, "proxy_user": proxyUser})
}
