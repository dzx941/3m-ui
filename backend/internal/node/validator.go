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

	// Config contains only node-level protocol options (for example TLS,
	// certificates, obfuscation, flow, and other Mihomo fields). Authentication
	// credentials are managed exclusively by ProxyUser/ListenerUser.
	if l.Config != "" {
		var custom map[string]interface{}
		if err := json.Unmarshal([]byte(l.Config), &custom); err != nil {
			return fmt.Errorf("invalid dynamic configuration parameters (must be valid JSON): %w", err)
		}
	}

	return nil
}
