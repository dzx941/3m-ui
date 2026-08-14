package listener

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
	"gopkg.in/yaml.v3"
)

// GenerateConfigYAML converts persisted listeners to Mihomo's native listeners
// schema. The same protocol boundary is used by the visual editor and client
// export so excluded local proxy endpoints can never leak into this generator.
func GenerateConfigYAML(dbListeners []models.Listener) (string, error) {
	listenersList := make([]map[string]interface{}, 0, len(dbListeners))

	for _, dl := range dbListeners {
		if !dl.Enabled {
			continue
		}

		protocol := strings.ToLower(strings.TrimSpace(dl.Protocol))
		if protocol == "" {
			protocol = strings.ToLower(strings.TrimSpace(dl.Type))
		}
		if !mihomoConfig.IsMihomoListenerProtocol(protocol) {
			return "", fmt.Errorf("unsupported Mihomo listener protocol %q", protocol)
		}
		if dl.Port < 1 || dl.Port > 65535 {
			return "", fmt.Errorf("listener %q has invalid port %d", dl.Name, dl.Port)
		}

		listen := strings.TrimSpace(dl.BindAddress)
		if listen == "" {
			listen = strings.TrimSpace(dl.Listen)
		}
		if listen == "" {
			listen = "0.0.0.0"
		}

		lm := map[string]interface{}{
			"name":   dl.Name,
			"type":   protocol,
			"listen": listen,
			"port":   dl.Port,
		}
		if dl.Proxy != "" {
			lm["proxy"] = dl.Proxy
		}
		if dl.Rule != "" {
			lm["rule"] = dl.Rule
		}
		if dl.UDP && listenerHasUDP(protocol) {
			lm["udp"] = true
		}

		if strings.TrimSpace(dl.Config) != "" {
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(dl.Config), &extra); err != nil {
				return "", fmt.Errorf("listener %q has invalid protocol config: %w", dl.Name, err)
			}

			normalizeListenerUsers(protocol, extra)

			for k, v := range extra {
				if k != "tls" && k != "udp" {
					lm[k] = v
				}
			}
		}

		listenersList = append(listenersList, lm)
	}

	yamlBytes, err := yaml.Marshal(&MihomoConfig{Listeners: listenersList})
	if err != nil {
		return "", fmt.Errorf("failed to marshal mihomo config to yaml: %w", err)
	}
	return string(yamlBytes), nil
}

// normalizeListenerUsers fixes old stored listener configs where ShadowQUIC
// users were stored as an object. Mihomo requires a list for these listeners.
func normalizeListenerUsers(protocol string, cfg map[string]interface{}) {
	if protocol != "shadowquic" && protocol != "trusttunnel" {
		return
	}

	users, ok := cfg["users"]
	if !ok || users == nil {
		return
	}

	if _, ok := users.([]interface{}); ok {
		return
	}

	object, ok := users.(map[string]interface{})
	if !ok {
		return
	}

	list := make([]interface{}, 0, len(object))
	for username, password := range object {
		list = append(list, map[string]interface{}{
			"username": username,
			"password": password,
		})
	}
	cfg["users"] = list
}

func listenerHasUDP(protocol string) bool {
	return protocol == "shadowsocks" || protocol == "snell"
}
