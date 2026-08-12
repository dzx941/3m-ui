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
	"github.com/dzx941/3m-ui/backend/internal/security"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

func ResolveServerAddress(cfg *config.Config, req *http.Request) string {
	if envURL := os.Getenv("PUBLIC_URL"); envURL != "" {
		return cleanURLHost(envURL)
	}
	if cfg != nil && cfg.Server.PublicURL != "" {
		return cleanURLHost(cfg.Server.PublicURL)
	}
	if req != nil && req.Host != "" {
		host := req.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			return h
		}
		return strings.Trim(host, "[]")
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
	if h, _, err := net.SplitHostPort(u); err == nil {
		u = h
	}
	u = strings.Trim(u, "[]")
	if u == "" || strings.ContainsAny(u, "\r\n/\\") {
		return "127.0.0.1"
	}
	return u
}

func GetSubscriptionURL(cfg *config.Config, req *http.Request, token string, target string) string {
	var base string
	if envURL := os.Getenv("PUBLIC_URL"); envURL != "" {
		base = envURL
	} else if cfg != nil && cfg.Server.PublicURL != "" {
		base = cfg.Server.PublicURL
	} else if req != nil && req.Host != "" {
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

func GenerateRawConfig(db *gorm.DB, token models.AccessToken, req *http.Request) ([]byte, error) {
	if db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	if token.ListenerID == 0 {
		return nil, fmt.Errorf("access token is not bound to a listener")
	}

	var listener models.Listener
	if err := db.First(&listener, token.ListenerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("listener not found")
		}
		return nil, fmt.Errorf("failed to fetch listener: %w", err)
	}
	if !listener.Enabled {
		return nil, fmt.Errorf("listener is disabled")
	}

	serverHost := ResolveServerAddress(config.GlobalConfig, req)
	proxy := map[string]interface{}{
		"name":   listener.Name,
		"type":   strings.ToLower(strings.TrimSpace(listener.Protocol)),
		"server": serverHost,
		"port":   listener.Port,
	}

	// Protocol-specific listener options are stored as JSON and copied into
	// the generated Mihomo proxy entry without dropping protocol-only fields.
	if listener.Config != "" {
		var options map[string]interface{}
		if err := json.Unmarshal([]byte(listener.Config), &options); err != nil {
			return nil, fmt.Errorf("invalid listener config for %q: %w", listener.Name, err)
		}
		for key, value := range options {
			if key == "name" || key == "type" || key == "server" || key == "port" {
				continue
			}
			proxy[key] = value
		}
	}

	if listener.UDP {
		proxy["udp"] = true
	}

	cfg := map[string]interface{}{
		"proxies": []interface{}{proxy},
		"proxy-groups": []interface{}{
			map[string]interface{}{
				"name":    "PROXY",
				"type":    "select",
				"proxies": []string{listener.Name},
			},
		},
		"rules": []string{"MATCH,PROXY"},
	}

	// Listener credentials may be stored encrypted in Config. Do not add a
	// second credential source here; Config remains the single listener source.
	_ = security.Decrypt

	return yaml.Marshal(cfg)
}
