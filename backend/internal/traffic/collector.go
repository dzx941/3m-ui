package traffic

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/mihomo/api"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type ConnectionCache struct {
	LastUpload   int64
	LastDownload int64
	LastSeen     time.Time
}

type Collector struct {
	mu         sync.Mutex
	configPath string
	connCache  map[string]*ConnectionCache
	service    *Service
}

func NewCollector(configPath string, service *Service) *Collector {
	return &Collector{
		configPath: configPath,
		connCache:  make(map[string]*ConnectionCache),
		service:    service,
	}
}

func (c *Collector) Collect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	db := database.GlobalDB
	if db == nil {
		return nil
	}

	baseURL, secret := getMihomoAPIConfig(c.configPath)
	client := api.NewClient(baseURL, secret)

	// Pull speed from /traffic
	trafficSnapshot, err := client.Traffic()
	if err != nil {
		_ = c.markAllOffline(db)
		c.service.Update(0, 0, 0, nil)
		return err
	}

	// Pull active connections
	connectionsResp, err := client.Connections()
	if err != nil {
		_ = c.markAllOffline(db)
		c.service.Update(0, 0, 0, nil)
		return err
	}

	var users []models.ProxyUser
	if err := db.Find(&users).Error; err != nil {
		return err
	}

	var listeners []models.Listener
	if err := db.Find(&listeners).Error; err != nil {
		return err
	}

	listenerMap := make(map[string]models.Listener)
	for _, l := range listeners {
		listenerMap[l.Name] = l
	}

	usernameMap := make(map[string]models.ProxyUser)
	uuidMap := make(map[string]models.ProxyUser)
	for _, u := range users {
		usernameMap[u.Username] = u
		if u.UUID != "" {
			uuidMap[u.UUID] = u
		}
	}

	now := time.Now()
	onlineUserIDs := make(map[uint]bool)
	userIncrementsUp := make(map[uint]int64)
	userIncrementsDown := make(map[uint]int64)

	currentConnIDs := make(map[string]bool)
	var apiConns []api.Connection

	for _, conn := range connectionsResp.Connections {
		currentConnIDs[conn.ID] = true
		apiConns = append(apiConns, conn)

		var diffUp, diffDown int64
		cache, exists := c.connCache[conn.ID]
		if exists {
			diffUp = conn.Upload - cache.LastUpload
			diffDown = conn.Download - cache.LastDownload
			if diffUp < 0 {
				diffUp = conn.Upload
			}
			if diffDown < 0 {
				diffDown = conn.Download
			}
		} else {
			diffUp = conn.Upload
			diffDown = conn.Download
		}

		c.connCache[conn.ID] = &ConnectionCache{
			LastUpload:   conn.Upload,
			LastDownload: conn.Download,
			LastSeen:     now,
		}

		inboundUser := conn.Metadata.InboundUser
		var mappedUser *models.ProxyUser

		// Mapping Priority 1: Match Username
		if u, ok := usernameMap[inboundUser]; ok {
			mappedUser = &u
		} else if u, ok := uuidMap[inboundUser]; ok {
			// Mapping Priority 2: Match UUID (for VMess, VLESS, TUIC)
			mappedUser = &u
		}

		if mappedUser != nil {
			onlineUserIDs[mappedUser.ID] = true
			if diffUp > 0 || diffDown > 0 {
				userIncrementsUp[mappedUser.ID] += diffUp
				userIncrementsDown[mappedUser.ID] += diffDown
			}
		}
	}

	// Clean up stale cache connections
	for id := range c.connCache {
		if !currentConnIDs[id] {
			delete(c.connCache, id)
		}
	}

	// Update traffic database values
	for userID, diffUp := range userIncrementsUp {
		diffDown := userIncrementsDown[userID]
		if diffUp > 0 || diffDown > 0 {
			err := db.Model(&models.ProxyUser{}).Where("id = ?", userID).
				Updates(map[string]any{
					"traffic_used":   gorm.Expr("traffic_used + ?", diffUp+diffDown),
					"upload_bytes":   gorm.Expr("upload_bytes + ?", diffUp),
					"download_bytes": gorm.Expr("download_bytes + ?", diffDown),
				}).Error
			if err != nil {
				log.Printf("Failed to update ProxyUser %d traffic: %v", userID, err)
			}

			// Save TrafficRecord only if there was actually positive traffic increment
			record := &models.TrafficRecord{
				ProxyUserID:   userID,
				UploadBytes:   diffUp,
				DownloadBytes: diffDown,
				Online:        true,
			}
			if err := db.Create(record).Error; err != nil {
				log.Printf("Failed to create TrafficRecord for user %d: %v", userID, err)
			}
		}
	}

	// Set Online/LastSeen state for all users
	for _, u := range users {
		isOnline := onlineUserIDs[u.ID]
		if isOnline {
			db.Model(&models.ProxyUser{}).Where("id = ?", u.ID).
				Updates(map[string]any{
					"online":    true,
					"last_seen": now,
				})
		} else {
			db.Model(&models.ProxyUser{}).Where("id = ?", u.ID).
				Updates(map[string]any{
					"online": false,
				})
		}
	}

	c.service.Update(connectionsResp.UploadTotal, connectionsResp.DownloadTotal, len(connectionsResp.Connections), apiConns)
	c.service.SetRates(trafficSnapshot.Up, trafficSnapshot.Down)

	// Trigger Enforcement limits check
	EnforceLimits(db, c.configPath)

	return nil
}

func (c *Collector) markAllOffline(db *gorm.DB) error {
	return db.Model(&models.ProxyUser{}).Where("online = ?", true).Update("online", false).Error
}

func getMihomoAPIConfig(configPath string) (string, string) {
	baseURL := "http://127.0.0.1:9090"
	secret := "3m-ui-default-secret-key"

	data, err := os.ReadFile(configPath)
	if err == nil {
		var cfg struct {
			ExternalController string `yaml:"external-controller"`
			Secret             string `yaml:"secret"`
		}
		if yaml.Unmarshal(data, &cfg) == nil {
			if cfg.ExternalController != "" {
				host := cfg.ExternalController
				if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
					baseURL = "http://" + host
				} else {
					baseURL = host
				}
			}
			if cfg.Secret != "" {
				secret = cfg.Secret
			}
		}
	}
	return baseURL, secret
}
