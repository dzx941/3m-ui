package credentials

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kazeyukiro/3m-ui/backend/internal/database/models"
)

// EnsureListenerCredentials adds a stable client credential to listeners that
// require one but do not have one yet. It deliberately has no dependency on
// node, user, or config packages so credential generation can be shared by
// validation and export without creating import cycles.
func EnsureListenerCredentials(l *models.Listener) error {
	if l == nil {
		return fmt.Errorf("listener is nil")
	}
	cfg, err := decodeConfig(l.Config)
	if err != nil {
		return err
	}
	proto := strings.ToLower(strings.TrimSpace(l.Protocol))
	if proto == "" {
		proto = strings.ToLower(strings.TrimSpace(l.Type))
	}
	if requiresUserCredentials(proto) && !hasExportCredentials(proto, cfg) {
		switch proto {
		case "vless", "vmess":
			uuid, err := randomUUID()
			if err != nil {
				return fmt.Errorf("generate client uuid: %w", err)
			}
			cfg["users"] = []interface{}{map[string]interface{}{"username": "client", "uuid": uuid}}
		case "trojan", "hysteria2", "anytls", "mieru", "shadowquic", "tuic":
			password, err := randomSecret(24)
			if err != nil {
				return fmt.Errorf("generate client credential: %w", err)
			}
			username := "client"
			if proto == "tuic" {
				username, err = randomUUID()
				if err != nil {
					return fmt.Errorf("generate TUIC client uuid: %w", err)
				}
			}
			cfg["users"] = map[string]interface{}{username: password}
		}
		encoded, err := json.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("encode listener credentials: %w", err)
		}
		l.Config = string(encoded)
	}
	return nil
}

func requiresUserCredentials(proto string) bool {
	switch strings.ToLower(proto) {
	case "vless", "vmess", "trojan", "hysteria2", "anytls", "mieru", "shadowquic", "tuic":
		return true
	default:
		return false
	}
}
func hasExportCredentials(proto string, cfg map[string]interface{}) bool {
	users, ok := cfg["users"]
	if !ok || users == nil {
		return false
	}
	switch strings.ToLower(proto) {
	case "hysteria2", "anytls", "mieru", "tuic":
		m, ok := users.(map[string]interface{})
		return ok && len(m) > 0
	default:
		list, ok := users.([]interface{})
		return ok && len(list) > 0
	}
}
func randomSecret(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func randomUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
func decodeConfig(raw string) (map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]interface{}{}, nil
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("invalid listener configuration: %w", err)
	}
	if cfg == nil {
		return map[string]interface{}{}, nil
	}
	return cfg, nil
}
