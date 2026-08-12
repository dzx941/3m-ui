package node

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
)

func ValidateNode(l *models.Listener) error {
	l.Name = strings.TrimSpace(l.Name)
	if l.Name == "" { return fmt.Errorf("node name cannot be empty") }

	proto := strings.ToLower(strings.TrimSpace(l.Protocol))
	if proto == "" { proto = strings.ToLower(strings.TrimSpace(l.Type)) }
	if proto == "" { return fmt.Errorf("node protocol cannot be empty") }

	validProtos := map[string]bool{
		"shadowsocks": true,
		"vmess": true,
		"vless": true,
		"trojan": true,
		"hysteria2": true,
		"tuic": true,
	}
	if !validProtos[proto] { return fmt.Errorf("unsupported listener protocol: %s", l.Protocol) }
	l.Protocol = proto
	l.Type = proto

	if l.Port < 1 || l.Port > 65535 { return fmt.Errorf("port number must be between 1 and 65535") }
	l.BindAddress = strings.TrimSpace(l.BindAddress)
	if l.BindAddress == "" { l.BindAddress = strings.TrimSpace(l.Listen) }
	if l.BindAddress == "" { l.BindAddress = "0.0.0.0" }

	configMap := map[string]interface{}{}
	if l.Config != "" {
		if err := json.Unmarshal([]byte(l.Config), &configMap); err != nil { return fmt.Errorf("invalid listener configuration (must be valid JSON): %w", err) }
	}

	if hasTLSEnabler(proto, configMap) { l.TLS = true }
	if l.TLS && !hasTLSEnabler(proto, configMap) { return fmt.Errorf("%s listener TLS requires certificate/private-key or a supported TLS alternative", proto) }
	return nil
}

func hasTLSEnabler(proto string, cfg map[string]interface{}) bool {
	cert := nonEmptyString(cfg["certificate"])
	key := nonEmptyString(cfg["private-key"])
	if key == "" { key = nonEmptyString(cfg["private_key"]) }
	if cert != "" && key != "" { return true }
	if proto != "vless" && proto != "trojan" { return false }
	if boolValue(cfg["allow-insecure"]) { return true }
	for _, key := range []string{"reality-config", "shadow-tls", "res-tls", "jls-config"} {
		if _, ok := cfg[key]; ok { return true }
	}
	if proto == "vless" { if _, ok := cfg["decryption"]; ok { return true } }
	if proto == "trojan" { if _, ok := cfg["ss-option"]; ok { return true } }
	return false
}

func nonEmptyString(v interface{}) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func boolValue(v interface{}) bool { b, _ := v.(bool); return b }
