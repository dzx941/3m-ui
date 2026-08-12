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

// ConfigEngine manages configuration generation, merging, and validation.
type ConfigEngine struct { db *gorm.DB }

func NewConfigEngine(db *gorm.DB) *ConfigEngine { return &ConfigEngine{db: db} }

func (ce *ConfigEngine) GenerateFinalConfig() (string, error) {
	baseBytes, err := yaml.Marshal(GetDefaultTemplate())
	if err != nil { return "", err }
	var merged map[string]interface{}
	if err := yaml.Unmarshal(baseBytes, &merged); err != nil { return "", err }

	var customFragments []models.Config
	if err := ce.db.Where("enabled = ?", true).Find(&customFragments).Error; err != nil {
		return "", fmt.Errorf("load custom config fragments: %w", err)
	}
	for _, fragment := range customFragments {
		var fragMap map[string]interface{}
		if err := yaml.Unmarshal([]byte(fragment.Content), &fragMap); err != nil {
			return "", fmt.Errorf("invalid custom config %q: %w", fragment.Name, err)
		}
		for k, v := range fragMap { merged[k] = v }
	}

	var listeners []models.Listener
	if err := ce.db.Where("enabled = ?", true).Find(&listeners).Error; err != nil {
		return "", fmt.Errorf("load enabled nodes: %w", err)
	}

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

// generateListeners emits the actual Mihomo `listeners` schema. In particular,
// server TLS uses certificate/private-key at the listener root; it is not a
// nested `tls: {enable: ...}` block. Authentication shapes also differ by
// protocol and are normalized here instead of relying on outbound proxy syntax.
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
			if err := json.Unmarshal([]byte(l.Config), &configMap); err != nil {
				return nil, fmt.Errorf("invalid listener config for %q: %w", l.Name, err)
			}
		}

		// The database has historically used both private_key and private-key.
		// Mihomo listener syntax only accepts the latter.
		if v, ok := configMap["certificate"]; ok { m["certificate"] = v }
		if v, ok := configMap["private-key"]; ok { m["private-key"] = v }
		if v, ok := configMap["private_key"]; ok { m["private-key"] = v }
		if l.TLS && (m["certificate"] == nil || m["private-key"] == nil) {
			// Do not emit an invalid nested TLS object. The validator is responsible
			// for rejecting a TLS-enabled listener without a valid TLS mechanism.
			delete(m, "certificate")
			delete(m, "private-key")
		}

		// Copy listener-side options while explicitly excluding outbound-only keys
		// that the old generator accidentally wrote into the server listener block.
		for k, v := range configMap {
			if !isCredentialOrOutboundOnly(k) && k != "private_key" && k != "private-key" && k != "certificate" {
				m[k] = v
			}
		}

		listenerCreds := creds[l.ID]
		switch protocol {
		case "shadowsocks":
			if len(listenerCreds) > 1 {
				return nil, fmt.Errorf("listener %q: shadowsocks supports one password; %d active proxy users are bound", l.Name, len(listenerCreds))
			}
			if cipher, ok := configMap["cipher"]; ok { m["cipher"] = cipher }
			if len(listenerCreds) == 1 {
				m["password"] = listenerCreds[0].Password
			} else if password, ok := configMap["password"]; ok {
				m["password"] = password
			}
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
			for _, cred := range listenerCreds {
				if cred.Username != "" { users[cred.Username] = cred.Password }
			}
			if len(users) > 0 { m["users"] = users } else if v, ok := configMap["users"]; ok { m["users"] = v }
		case "tuic":
			users := make(map[string]string)
			for _, cred := range listenerCreds {
				if cred.UUID != "" { users[cred.UUID] = cred.Password }
			}
			if len(users) > 0 { m["users"] = users } else if v, ok := configMap["users"]; ok { m["users"] = v }
		}

		result = append(result, m)
	}
	return result, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" { return strings.TrimSpace(value) }
	}
	return ""
}

func isCredentialOrOutboundOnly(key string) bool {
	switch key {
	case "password", "username", "uuid", "flow", "users",
		"tls", "sni", "servername", "skip-cert-verify", "name-cert-verify",
		"fingerprint", "client-fingerprint", "encryption", "alterId",
		"network", "ws-opts", "grpc-opts", "h2-opts", "http-opts", "mkcp-opts",
		"reality-opts", "shadow-tls-opts", "restls-opts", "jls-opts",
		"certificate", "private-key", "private_key":
		return true
	default:
		return false
	}
}
