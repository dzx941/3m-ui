package traffic

import (
	"net/http"
	"strconv"

	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.GetSnapshot())
}

func (h *Handler) Users(c *gin.Context) {
	db := database.GlobalDB
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	var users []models.ProxyUser
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res := make([]UserTraffic, 0, len(users))
	for _, u := range users {
		res = append(res, UserTraffic{
			UserID:        u.ID,
			Username:      u.Username,
			UploadBytes:   u.UploadBytes,
			DownloadBytes: u.DownloadBytes,
			TrafficUsed:   u.TrafficUsed,
			TrafficLimit:  u.TrafficLimit,
			Online:        u.Online,
		})
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) Connections(c *gin.Context) {
	apiConns := h.service.GetConnections()

	flatConns := make([]map[string]any, 0, len(apiConns))
	for _, conn := range apiConns {
		dstPort, _ := strconv.Atoi(conn.Metadata.DestinationPort)

		flat := map[string]any{
			"id":               conn.ID,
			"network":          conn.Metadata.Network,
			"upload":           conn.Upload,
			"download":         conn.Download,
			"source":           conn.Metadata.SourceIP,
			"destination":      conn.Metadata.DestinationIP + ":" + conn.Metadata.DestinationPort,
			"host":             conn.Metadata.Host,
			"destination_port": dstPort,
			"inbound_name":     conn.Metadata.InboundName,
			"inbound_user":     conn.Metadata.InboundUser,
		}
		flatConns = append(flatConns, flat)
	}

	c.JSON(http.StatusOK, gin.H{"connections": flatConns})
}
