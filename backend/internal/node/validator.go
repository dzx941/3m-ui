package node

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
)

// ValidateNode performs strict validation checks on a node model
func ValidateNode(l *models.Listener) error {
	l.Name = strings.TrimSpace(l.Name)
	if l.Name == "" {
		return fmt.Errorf("node name cannot be empty")
	}

	proto := strings.ToLower(l.Protocol)
	if proto == "" {
		return fmt.Errorf("node protocol cannot be empty")
	}

	validProtos := map[string]bool{
		"shadowsocks": true,
		"vmess":       true,
		"vless":       true,
		"trojan":      true,
		"hysteria2":   true,
		"tuic":        true,
	}
	if !validProtos[proto] {
		return fmt.Errorf("unsupported node protocol: %s", l.Protocol)
	}

	if l.Port < 1 || l.Port > 65535 {
		return fmt.Errorf("port number must be between 1 and 65535")
	}

	l.BindAddress = strings.TrimSpace(l.BindAddress)
	if l.BindAddress == "" {
		l.BindAddress = "0.0.0.0"
	}

	// Validate config fields by trying to unmarshal
	if l.Config != "" {
		var custom map[string]interface{}
		if err := json.Unmarshal([]byte(l.Config), &custom); err != nil {
			return fmt.Errorf("invalid dynamic configuration parameters (must be a valid JSON): %w", err)
		}

		// Perform specific protocol checks
		switch proto {
		case "shadowsocks":
			if _, exists := custom["password"]; !exists {
				return fmt.Errorf("shadowsocks protocol requires a 'password' configuration parameter")
			}
		case "vmess":
			if _, exists := custom["uuid"]; !exists {
				return fmt.Errorf("vmess protocol requires a 'uuid' configuration parameter")
			}
		case "vless":
			if _, exists := custom["uuid"]; !exists {
				return fmt.Errorf("vless protocol requires a 'uuid' configuration parameter")
			}
		case "trojan":
			if _, exists := custom["password"]; !exists {
				return fmt.Errorf("trojan protocol requires a 'password' configuration parameter")
			}
		}
	} else {
		// If empty, return warning depending on protocol
		if proto == "shadowsocks" || proto == "vmess" || proto == "vless" || proto == "trojan" {
			return fmt.Errorf("%s protocol requires credentials configured in JSON", proto)
		}
	}

	return nil
}
