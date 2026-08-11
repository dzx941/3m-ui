package converter

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// ResolveServerAddress cleans and returns the preferred server Host IP/Domain.
func ResolveServerAddress(cfg *config.Config, req *http.Request) string {
	if envURL := os.Getenv("PUBLIC_URL"); envURL != "" {
		return cleanURLHost(envURL)
	}
	if cfg != nil && cfg.Server.PublicURL != "" {
		return cleanURLHost(cfg.Server.PublicURL)
	}
	if req != nil && req.Host != "" {
		host := req.Host
		if strings.Contains(host, ":") {
			h, _, err := net.SplitHostPort(host)
			if err == nil {
				return h
			}
		}
		return host
	}
	return "127.0.0.1"
}

func cleanURLHost(u string) string {
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	u = strings.Split(u, "/")[0]
	if strings.Contains(u, ":") {
		h, _, err := net.SplitHostPort(u)
		if err == nil {
			return h
		}
	}
	return u
}

// GetSubscriptionURL builds the final client subscription URL.
func GetSubscriptionURL(cfg *config.Config, req *http.Request, token string, target string) string {
	var base string
	if envURL := os.Getenv("PUBLIC_URL"); envURL != "" {
		base = envURL
	} else if cfg != nil && cfg.Server.PublicURL != "" {
		base = cfg.Server.PublicURL
	} else if req != nil {
		scheme := "http"
		if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		base = fmt.Sprintf("%s://%s", scheme, req.Host)
	} else {
		base = "http://127.0.0.1:8080"
	}
	base = strings.TrimSuffix(base, "/")
	if target == "" {
		return fmt.Sprintf("%s/api/v1/client/sub/%s", base, token)
	}
	return fmt.Sprintf("%s/api/v1/client/sub/%s?target=%s", base, token, target)
}

// GenerateRawConfig produces a standard Clash YAML configuration for a given token.
func GenerateRawConfig(db *gorm.DB, token models.AccessToken, req *http.Request) ([]byte, error) {
	proxies := make([]map[string]interface{}, 0)

	serverHost := ResolveServerAddress(config.GlobalConfig, req)

	if token.Type == "user" {
		// Fetch listeners assigned to this proxy user
		var listeners []models.Listener
		err := db.Raw(`
			SELECT listeners.* FROM listeners
			JOIN listener_users ON listener_users.listener_id = listeners.id
			WHERE listener_users.proxy_user_id = ?
			AND listeners.enabled = 1
		`, token.TargetID).Scan(&listeners).Error
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user listeners: %w", err)
		}

		// Also fetch proxy user credentials
		var u models.ProxyUser
		if err := db.First(&u, token.TargetID).Error; err == nil {
			for _, l := range listeners {
				p := map[string]interface{}{
					"name":   l.Name,
					"type":   l.Protocol,
					"server": serverHost,
					"port":   l.Port,
					"udp":    l.UDP,
				}

				// Map protocol options from JSON config
				var opts map[string]interface{}
				if err := json.Unmarshal([]byte(l.Config), &opts); err == nil {
					for k, v := range opts {
						p[k] = v
					}
				}

				// Populate standard credentials if not already present
				if _, ok := p["password"]; !ok && u.Username != "" {
					p["password"] = u.Username
				}
				if _, ok := p["uuid"]; !ok && u.Username != "" {
					p["uuid"] = u.Username
				}

				proxies = append(proxies, p)
			}
		}
	} else if token.Type == "proxy" {
		// Fetch specific external ProxyNode
		visual, err := mihomoConfig.GetVisualConfig(db)
		if err != nil {
			return nil, fmt.Errorf("failed to load visual config: %w", err)
		}
		idx := int(token.TargetID)
		if idx >= 0 && idx < len(visual.Proxies) {
			pe := visual.Proxies[idx]
			p := map[string]interface{}{
				"name":   pe.Name,
				"type":   pe.Type,
				"server": pe.Server,
				"port":   pe.Port,
			}
			for k, v := range pe.Options {
				p[k] = v
			}
			proxies = append(proxies, p)
		}
	}

	cfg := map[string]interface{}{
		"mixed-port": 7890,
		"proxies":    proxies,
		"proxy-groups": []interface{}{
			map[string]interface{}{
				"name":    "PROXY",
				"type":    "select",
				"proxies": getProxyNames(proxies),
			},
		},
		"rules": []string{
			"MATCH,PROXY",
		},
	}

	return yaml.Marshal(cfg)
}

func getProxyNames(proxies []map[string]interface{}) []string {
	names := make([]string, 0, len(proxies))
	for _, p := range proxies {
		if name, ok := p["name"].(string); ok {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		names = append(names, "DIRECT")
	}
	return names
}
