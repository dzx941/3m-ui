package converter

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
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

func cleanURLHost(raw string) string {
	raw = strings.TrimSpace(raw)
	u := raw
	if parsed, err := url.Parse(raw); err == nil && parsed.Host != "" {
		u = parsed.Host
	} else {
		u = strings.TrimPrefix(u, "https://")
		u = strings.TrimPrefix(u, "http://")
		u = strings.Split(u, "/")[0]
	}
	if strings.Contains(u, ":") {
		h, _, err := net.SplitHostPort(u)
		if err == nil {
			return h
		}
	}
	if u == "" || strings.ContainsAny(u, "\r\n/\\") {
		return "127.0.0.1"
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
	} else if req != nil && req.Host != "" {
		// Do not trust a client-supplied X-Forwarded-Proto header. Deployments
		// behind TLS-terminating proxies should set PUBLIC_URL/public_url.
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}
		base = fmt.Sprintf("%s://%s", scheme, req.Host)
	} else {
		base = "http://127.0.0.1:8080"
	}
	base = strings.TrimSuffix(base, "/")
	pathToken := url.PathEscape(token)
	if target == "" {
		return fmt.Sprintf("%s/api/v1/client/sub/%s", base, pathToken)
	}
	return fmt.Sprintf("%s/api/v1/client/sub/%s?target=%s", base, pathToken, url.QueryEscape(strings.ToLower(strings.TrimSpace(target))))
}

// GenerateRawConfig produces a standard Clash YAML configuration for a given token.
func GenerateRawConfig(db *gorm.DB, token models.AccessToken, req *http.Request) ([]byte, error) {
	if db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	if token.Type != "user" && token.Type != "proxy" {
		return nil, fmt.Errorf("invalid access token type")
	}

	proxies := make([]map[string]interface{}, 0)
	serverHost := ResolveServerAddress(config.GlobalConfig, req)

	if token.Type == "user" {
		var u models.ProxyUser
		if err := db.First(&u, token.TargetID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("proxy user not found")
			}
			return nil, fmt.Errorf("failed to fetch proxy user: %w", err)
		}

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

		for _, l := range listeners {
			p := map[string]interface{}{
				"name":   l.Name,
				"type":   l.Protocol,
				"server": serverHost,
				"port":   l.Port,
				"udp":    l.UDP,
			}

			var opts map[string]interface{}
			if l.Config != "" {
				if err := json.Unmarshal([]byte(l.Config), &opts); err != nil {
					return nil, fmt.Errorf("invalid listener config for %q: %w", l.Name, err)
				}
				for k, v := range opts {
					p[k] = v
				}
			}

			if _, ok := p["password"]; !ok && u.Username != "" {
				p["password"] = u.Username
			}
			if _, ok := p["uuid"]; !ok && u.Username != "" {
				p["uuid"] = u.Username
			}
			proxies = append(proxies, p)
		}
	} else {
		visual, err := mihomoConfig.GetVisualConfig(db)
		if err != nil {
			return nil, fmt.Errorf("failed to load visual config: %w", err)
		}
		idx := int(token.TargetID)
		if idx < 0 || idx >= len(visual.Proxies) {
			return nil, fmt.Errorf("proxy node not found")
		}
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
		"rules": []string{"MATCH,PROXY"},
	}
	return yaml.Marshal(cfg)
}

func getProxyNames(proxies []map[string]interface{}) []string {
	names := make([]string, 0, len(proxies))
	for _, p := range proxies {
		if name, ok := p["name"].(string); ok && strings.TrimSpace(name) != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		names = append(names, "DIRECT")
	}
	return names
}
