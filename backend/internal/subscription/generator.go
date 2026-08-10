package subscription

import (
	"encoding/base64"
	"encoding/json"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type Generator struct {
	db *gorm.DB
}

func NewGenerator(db *gorm.DB) *Generator {
	return &Generator{db: db}
}

type ClientProxy struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Server   string `yaml:"server,omitempty"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password,omitempty"`
	UUID     string `yaml:"uuid,omitempty"`
	TLS      bool   `yaml:"tls,omitempty"`
	UDP      bool   `yaml:"udp,omitempty"`
}

func (g *Generator) Generate(userID uint) ([]byte, error) {
	var listeners []models.Listener

	err := g.db.Raw(`
		SELECT listeners.* FROM listeners
		JOIN listener_users ON listener_users.listener_id = listeners.id
		WHERE listener_users.proxy_user_id = ?
		AND listeners.enabled = 1
	`, userID).Scan(&listeners).Error
	if err != nil {
		return nil, err
	}

	proxies := make([]ClientProxy, 0)
	for _, l := range listeners {
		var extra map[string]interface{}
		_ = json.Unmarshal([]byte(l.Config), &extra)

		p := ClientProxy{
			Name: l.Name,
			Type: l.Protocol,
			Server: l.BindAddress,
			Port: l.Port,
			TLS: l.TLS,
			UDP: l.UDP,
		}

		if v, ok := extra["password"].(string); ok {
			p.Password = v
		}
		if v, ok := extra["uuid"].(string); ok {
			p.UUID = v
		}

		proxies = append(proxies, p)
	}

	cfg := map[string]interface{}{
		"mixed-port": 7890,
		"proxies": proxies,
		"proxy-groups": []interface{}{
			map[string]interface{}{
				"name": "AUTO",
				"type": "select",
				"proxies": []string{"DIRECT"},
			},
		},
	}

	return yaml.Marshal(cfg)
}

func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
