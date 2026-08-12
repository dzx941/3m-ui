package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type ConfigEngine struct { db *gorm.DB }

func NewConfigEngine(db *gorm.DB) *ConfigEngine { return &ConfigEngine{db: db} }

func (ce *ConfigEngine) GenerateFinalConfig() (string, error) {
	baseBytes, err := yaml.Marshal(GetDefaultTemplate())
	if err != nil { return "", err }
	var merged map[string]interface{}
	if err := yaml.Unmarshal(baseBytes, &merged); err != nil { return "", err }

	var customFragments []models.Config
	if err := ce.db.Where("enabled = ?", true).Find(&customFragments).Error; err != nil { return "", fmt.Errorf("load custom config fragments: %w", err) }
	for _, fragment := range customFragments {
		var fragMap map[string]interface{}
		if err := yaml.Unmarshal([]byte(fragment.Content), &fragMap); err != nil { return "", fmt.Errorf("invalid custom config %q: %w", fragment.Name, err) }
		for k, v := range fragMap { merged[k] = v }
	}

	var listeners []models.Listener
	if err := ce.db.Where("enabled = ?", true).Find(&listeners).Error; err != nil { return "", fmt.Errorf("load enabled nodes: %w", err) }

	credentials := make(map[uint][]user.Credential)
	if user.GlobalService != nil {
		credentials, err = user.GlobalService.ActiveCredentialsByListener()
		if err != nil { return "", fmt.Errorf("load proxy user credentials: %w", err) }
	}

	generated, err := generateListeners(listeners, credentials)
	if err != nil { return "", err }
	merged["listeners"] = generated

	finalBytes, err := yaml.Marshal(merged)
	if err != nil { return "", fmt.Errorf("serialize final configuration: %w", err) }
	return string(finalBytes), nil
}

// generateListeners emits the actual Mihomo `listeners` schema. Server TLS is
// represented by listener-level certificate/private-key fields, while client
// TLS fields such as `tls: true` are deliberately excluded from this block.
func generateListeners(listeners []models.Listener, creds map[uint][]user.Credential) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(listeners))
	for _, l := range listeners {
		protocol := strings.ToLower(strings.TrimSpace(l.Protocol))
		if protocol == "" { protocol = strings.ToLower(strings.TrimSpace(l.Type)) }

		m := map[string]interface{}{
			"name": l.Name,
			"type": protocol,
			"port": l.Port,
			"listen": firstNonEmpty(l.BindAddress, l.Listen, "0.0.0.0"),
		}
		if l.Proxy != "" { m["proxy"] = l.Proxy }
		if l.Rule != "" { m["rule"] = l.Rule }
		if l.UDP { m["udp"] = true }

		configMap := map[string]interface{}{}
		if l.Config != "" {
			if err := json.Unmarshal([]byte(l.Config), &configMap); err != nil { return nil, fmt.Errorf("invalid listener config for %q: %w", l.Name, err) }
		}

		if v, ok := configMap["certificate"]; ok { m["certificate"] = v }
		if v, ok := configMap["private-key"]; ok { m["private-key"] = v }
		if v, ok := configMap["private_key"]; ok { m["private-key"] = v }
		if l.TLS && !hasServerTLS(protocol, configMap) {
			return nil, fmt.Errorf("listener %q: TLS is enabled but no valid server TLS mechanism is configured", l.Name)
		}

		for k, v := range configMap {
			if !isCredentialOrOutboundOnly(k) && k != "private_key" && k != "private-key" && k != "certificate" { m[k] = v }
		}

		listenerCreds := creds[l.ID]
		switch protocol {
		case "shadowsocks":
			if len(listenerCreds) > 1 { return nil, fmt.Errorf("listener %q: shadowsocks supports one password; %d active proxy users are bound", l.Name, len(listenerCreds)) }
			if cipher, ok := configMap["cipher"]; ok { m["cipher"] = cipher }
			if len(listenerCreds) == 1 { m["password"] = listenerCreds[0].Password } else if password, ok := configMap["password"]; ok { m["password"] = password }
		case "vmess":
			users := make([]map[string]interface{}, 0, len(listenerCreds))
			for _, cred := range listenerCreds {
				u := map[string]interface{}{"uuid": cred.UUID}
				if cred.Username != "" { u["username"] = cred.Username }
				if alterID, ok := configMap["alterId"]; ok { u["alterId"] = alterID }
				users = append(users, u)
			}
			if len(users) > 0 { m["users"] = users } else if v, ok := configMap["users"]; ok { m["users"] = v }
		case "vless":
			users := make([]map[string]interface{}, 0, len(listenerCreds))
			for _, cred := range listenerCreds {
				u := map[string]interface{}{"uuid": cred.UUID}
				if cred.Username != "" { u["username"] = cred.Username }
				if flow, ok := configMap["flow"]; ok { u["flow"] = flow }
				users = append(users, u)
			}
			if len(users) > 0 { m["users"] = users } else if v, ok := configMap["users"]; ok { m["users"] = v }
		case "trojan":
			users := make([]map[string]interface{}, 0, len(listenerCreds))
			for _, cred := range listenerCreds {
				u := map[string]interface{}{"password": cred.Password}
				if cred.Username != "" { u["username"] = cred.Username }
				users = append(users, u)
			}
			if len(users) > 0 { m["users"] = users } else if v, ok := configMap["users"]; ok { m["users"] = v }
		case "hysteria2":
			users := make(map[string]string)
			for _, cred := range listenerCreds { if cred.Username != "" { users[cred.Username] = cred.Password } }
			if len(users) > 0 { m["users"] = users } else if v, ok := configMap["users"]; ok { m["users"] = v }
		case "tuic":
			users := make(map[string]string)
			for _, cred := range listenerCreds { if cred.UUID != "" { users[cred.UUID] = cred.Password } }
			if len(users) > 0 { m["users"] = users } else if v, ok := configMap["users"]; ok { m["users"] = v }
		}

		result = append(result, m)
	}
	return result, nil
}

func hasServerTLS(protocol string, cfg map[string]interface{}) bool {
	cert := nonEmptyString(cfg["certificate"])
	key := nonEmptyString(cfg["private-key"])
	if key == "" { key = nonEmptyString(cfg["private_key"]) }
	if cert != "" && key != "" { return true }
	if protocol != "vless" && protocol != "trojan" { return false }
	if boolValue(cfg["allow-insecure"]) { return true }
	for _, name := range []string{"reality-config", "shadow-tls", "res-tls", "jls-config"} {
		if _, ok := cfg[name]; ok { return true }
	}
	if protocol == "vless" { if _, ok := cfg["decryption"]; ok { return true } }
	if protocol == "trojan" { if _, ok := cfg["ss-option"]; ok { return true } }
	return false
}

func nonEmptyString(v interface{}) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func boolValue(v interface{}) bool { b, _ := v.(bool); return b }

func firstNonEmpty(values ...string) string {
	for _, value := range values { if strings.TrimSpace(value) != "" { return strings.TrimSpace(value) } }
	return ""
}

func isCredentialOrOutboundOnly(key string) bool {
	switch key {
	case "password", "username", "uuid", "flow", "users", "tls", "sni", "servername", "skip-cert-verify", "name-cert-verify", "fingerprint", "client-fingerprint", "encryption", "alterId", "network", "ws-opts", "grpc-opts", "h2-opts", "http-opts", "mkcp-opts", "reality-opts", "shadow-tls-opts", "restls-opts", "jls-opts", "certificate", "private-key", "private_key":
		return true
	default:
		return false
	}
}
