package listener

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"gopkg.in/yaml.v3"
)

var supportedProtocols = map[string]struct{}{
	"socks": {},
	"http": {},
	"tproxy": {},
	"redir": {},
	"mixed": {},
	"tunnel": {},
	"tun": {},
	"shadowsocks": {},
	"snell": {},
	"vmess": {},
	"vless": {},
	"trojan": {},
	"hysteria2": {},
	"hysteria2-realm": {},
	"tuic": {},
	"shadowquic": {},
	"anytls": {},
	"mieru": {},
	"sudoku": {},
	"trusttunnel": {},
}

// SupportedProtocols returns every listener type accepted by Mihomo v1.19.x.
func SupportedProtocols() []string {
	return []string{
		"socks", "http", "tproxy", "redir", "mixed", "tunnel", "tun",
		"shadowsocks", "snell", "vmess", "vless", "trojan", "hysteria2",
		"hysteria2-realm", "tuic", "shadowquic", "anytls", "mieru", "sudoku", "trusttunnel",
	}
}

func normalizeProtocol(dl models.Listener) string {
	protocol := strings.ToLower(strings.TrimSpace(dl.Protocol))
	if protocol == "" {
		protocol = strings.ToLower(strings.TrimSpace(dl.Type))
	}
	return protocol
}

// GenerateConfigYAML converts persisted listeners to the exact top-level
// Mihomo listeners schema. Protocol-specific fields remain in Config so that
// every protocol can retain its own native options without lossy conversion.
func GenerateConfigYAML(dbListeners []models.Listener) (string, error) {
	listenersList := make([]map[string]interface{}, 0, len(dbListeners))

	for _, dl := range dbListeners {
		if !dl.Enabled {
			continue
		}

		protocol := normalizeProtocol(dl)
		if _, ok := supportedProtocols[protocol]; !ok {
			return "", fmt.Errorf("unsupported Mihomo listener protocol %q; supported protocols: %s", protocol, strings.Join(SupportedProtocols(), ", "))
		}

		listen := strings.TrimSpace(dl.BindAddress)
		if listen == "" {
			listen = strings.TrimSpace(dl.Listen)
		}
		if listen == "" {
			listen = "0.0.0.0"
		}
		if dl.Port < 1 || dl.Port > 65535 {
			return "", fmt.Errorf("listener %q has invalid port %d", dl.Name, dl.Port)
		}

		lm := map[string]interface{}{
			"name":   dl.Name,
			"type":   protocol,
			"listen": listen,
			"port":   dl.Port,
		}

		if dl.UDP && protocolSupportsUDP(protocol) {
			lm["udp"] = true
		}
		if dl.Proxy != "" {
			lm["proxy"] = dl.Proxy
		}
		if dl.Rule != "" {
			lm["rule"] = dl.Rule
		}

		if dl.Config != "" {
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(dl.Config), &extra); err != nil {
				if yamlErr := yaml.Unmarshal([]byte(dl.Config), &extra); yamlErr != nil {
					return "", fmt.Errorf("listener %q has invalid protocol config: %w", dl.Name, err)
				}
			}
			for k, v := range extra {
				if strings.TrimSpace(k) != "" {
					lm[k] = v
				}
			}
		}

		listenersList = append(listenersList, lm)
	}

	cfg := MihomoConfig{Listeners: listenersList}
	yamlBytes, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal mihomo config to yaml: %w", err)
	}
	return string(yamlBytes), nil
}

func protocolSupportsUDP(protocol string) bool {
	switch protocol {
	case "socks", "tproxy", "mixed", "shadowsocks", "snell", "hysteria2", "hysteria2-realm", "tuic", "shadowquic", "mieru", "sudoku", "trusttunnel":
		return true
	default:
		return false
	}
}
