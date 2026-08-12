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

type ConfigEngine struct {
	db *gorm.DB
}

func NewConfigEngine(db *gorm.DB) *ConfigEngine { return &ConfigEngine{db: db} }

func (ce *ConfigEngine) GenerateFinalConfig() (string, error) {
	baseBytes, err := yaml.Marshal(GetDefaultTemplate())
	if err != nil {
		return "", err
	}
	var merged map[string]interface{}
	if err := yaml.Unmarshal(baseBytes, &merged); err != nil {
		return "", err
	}

	var customFragments []models.Config
	if err := ce.db.Where("enabled = ?", true).Find(&customFragments).Error; err != nil {
		return "", fmt.Errorf("load custom config fragments: %w", err)
	}
	for _, fragment := range customFragments {
		var fragMap map[string]interface{}
		if err := yaml.Unmarshal([]byte(fragment.Content), &fragMap); err != nil {
			return "", fmt.Errorf("invalid custom config %q: %w", fragment.Name, err)
		}
		for k, v := range fragMap {
			merged[k] = v
		}
	}

	var listeners []models.Listener
	if err := ce.db.Where("enabled = ?", true).Find(&listeners).Error; err != nil {
		return "", fmt.Errorf("load enabled listeners: %w", err)
	}

	credentials := make(map[uint][]user.Credential)
	if user.GlobalService != nil {
		credentials, err = user.GlobalService.ActiveCredentialsByListener()
		if err != nil {
			return "", fmt.Errorf("load proxy user credentials: %w", err)
		}
	}

	generated, err := generateListeners(listeners, credentials)
	if err != nil {
		return "", err
	}
	merged["listeners"] = generated

	finalBytes, err := yaml.Marshal(merged)
	if err != nil {
		return "", fmt.Errorf("serialize final configuration: %w", err)
	}
	return string(finalBytes), nil
}

// generateListeners emits Mihomo's real listeners schema. The UI deliberately
// exposes only listener protocols that have a corresponding distributable
// client proxy configuration. Local proxy ports, TUN and WireGuard are not
// listener/client distribution protocols.
func generateListeners(listeners []models.Listener, creds map[uint][]user.Credential) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(listeners))
	for _, l := range listeners {
		protocol := strings.ToLower(strings.TrimSpace(l.Protocol))
		if protocol == "" {
			protocol = strings.ToLower(strings.TrimSpace(l.Type))
		}
		if !IsMihomoListenerProtocol(protocol) {
			return nil, fmt.Errorf("unsupported Mihomo listener protocol %q", protocol)
		}
		if protocol == "hysteria2-realm" {
			// This is an auxiliary HTTP/HTTPS realm service, not a client proxy
			// endpoint. It remains valid in Mihomo, but is not distributed as a
			// client node by 3m-ui.
		}
		if l.Port < 1 || l.Port > 65535 {
			return nil, fmt.Errorf("listener %q has invalid port %d", l.Name, l.Port)
		}

		listen := strings.TrimSpace(l.BindAddress)
		if listen == "" {
			listen = strings.TrimSpace(l.Listen)
		}
		if listen == "" {
			listen = "0.0.0.0"
		}

		m := map[string]interface{}{
			"name":   l.Name,
			"type":   protocol,
			"port":   l.Port,
			"listen": listen,
		}

		configMap, err := decodeListenerConfig(l.Config)
		if err != nil {
			return nil, fmt.Errorf("listener %q: %w", l.Name, err)
		}

		if l.Proxy != "" {
			m["proxy"] = l.Proxy
		}
		if l.Rule != "" {
			m["rule"] = l.Rule
		}
		if value, ok := configMap["routing-mark"]; ok {
			m["routing-mark"] = value
		}
		if l.UDP && listenerHasUDPOption(protocol) {
			m["udp"] = true
		}

		// Keep native listener fields. Only client-only fields and credentials
		// that are rebuilt below are removed.
		for k, v := range configMap {
			if listenerFieldIsManaged(k) {
				continue
			}
			m[k] = v
		}

		copyServerTLSFields(m, configMap)
		listenerCreds := creds[l.ID]

		switch protocol {
		case "shadowsocks":
			if len(listenerCreds) > 1 {
				return nil, fmt.Errorf("listener %q: Shadowsocks supports one password; %d active users are bound", l.Name, len(listenerCreds))
			}
			if value, ok := configMap["cipher"]; ok {
				m["cipher"] = value
			}
			if len(listenerCreds) == 1 && listenerCreds[0].Password != "" {
				m["password"] = listenerCreds[0].Password
			} else if value, ok := configMap["password"]; ok {
				m["password"] = value
			}
		case "snell":
			copyOption(m, configMap, "psk")
			copyOption(m, configMap, "version")
		case "vmess":
			users := make([]map[string]interface{}, 0, len(listenerCreds))
			for _, cred := range listenerCreds {
				if cred.UUID == "" {
					continue
				}
				u := map[string]interface{}{"uuid": cred.UUID}
				if cred.Username != "" {
					u["username"] = cred.Username
				}
				if alterID, ok := configMap["alterId"]; ok {
					u["alterId"] = alterID
				}
				users = append(users, u)
			}
			if len(users) > 0 {
				m["users"] = users
			} else if value, ok := configMap["users"]; ok {
				m["users"] = value
			}
		case "vless":
			users := make([]map[string]interface{}, 0, len(listenerCreds))
			for _, cred := range listenerCreds {
				if cred.UUID == "" {
					continue
				}
				u := map[string]interface{}{"uuid": cred.UUID}
				if cred.Username != "" {
					u["username"] = cred.Username
				}
				if flow, ok := configMap["flow"]; ok {
					u["flow"] = flow
				}
				users = append(users, u)
			}
			if len(users) > 0 {
				m["users"] = users
			} else if value, ok := configMap["users"]; ok {
				m["users"] = value
			}
		case "trojan":
			users := make([]map[string]interface{}, 0, len(listenerCreds))
			for _, cred := range listenerCreds {
				if cred.Password == "" {
					continue
				}
				u := map[string]interface{}{"password": cred.Password}
				if cred.Username != "" {
					u["username"] = cred.Username
				}
				users = append(users, u)
			}
			if len(users) > 0 {
				m["users"] = users
			} else if value, ok := configMap["users"]; ok {
				m["users"] = value
			}
		case "hysteria2", "anytls", "mieru":
			users := make(map[string]string)
			for _, cred := range listenerCreds {
				if cred.Username != "" && cred.Password != "" {
					users[cred.Username] = cred.Password
				}
			}
			if len(users) > 0 {
				m["users"] = users
			} else if value, ok := configMap["users"]; ok {
				m["users"] = value
			}
		case "tuic":
			// TUIC V4 uses token; TUIC V5 uses UUID/password users. Both
			// share type: tuic in Mihomo.
			if value, ok := configMap["token"]; ok {
				m["token"] = value
			} else {
				users := make(map[string]string)
				for _, cred := range listenerCreds {
					if cred.UUID != "" && cred.Password != "" {
						users[cred.UUID] = cred.Password
					}
				}
				if len(users) > 0 {
					m["users"] = users
				} else if value, ok := configMap["users"]; ok {
					m["users"] = value
				}
			}
		case "shadowquic", "trusttunnel":
			users := make([]map[string]interface{}, 0, len(listenerCreds))
			for _, cred := range listenerCreds {
				if cred.Username == "" || cred.Password == "" {
					continue
				}
				users = append(users, map[string]interface{}{"username": cred.Username, "password": cred.Password})
			}
			if len(users) > 0 {
				m["users"] = users
			} else if value, ok := configMap["users"]; ok {
				m["users"] = value
			}
		case "sudoku":
			// Sudoku is entirely key/config based; its native fields are copied.
		case "hysteria2-realm":
			// No user credentials are required for the realm service.
		}

		result = append(result, m)
	}
	return result, nil
}

func decodeListenerConfig(raw string) (map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]interface{}{}, nil
	}
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &configMap); err != nil {
		return nil, fmt.Errorf("invalid listener configuration (must be valid JSON): %w", err)
	}
	if configMap == nil {
		configMap = map[string]interface{}{}
	}
	return configMap, nil
}

func listenerFieldIsManaged(key string) bool {
	switch key {
	case "users", "username", "password", "uuid", "flow", "alterId",
		"tls", "servername", "sni", "skip-cert-verify", "name-cert-verify",
		"fingerprint", "client-fingerprint", "reality-opts", "shadow-tls-opts",
		"restls-opts", "jls-opts", "ws-opts", "grpc-opts", "h2-opts", "http-opts",
		"mkcp-opts", "certificate", "private-key", "private_key":
		return true
	default:
		return false
	}
}

func copyServerTLSFields(dst, src map[string]interface{}) {
	if value, ok := src["certificate"]; ok {
		dst["certificate"] = value
	}
	if value, ok := src["private-key"]; ok {
		dst["private-key"] = value
	} else if value, ok := src["private_key"]; ok {
		dst["private-key"] = value
	}
}

func copyOption(dst, src map[string]interface{}, key string) {
	if value, ok := src[key]; ok {
		dst[key] = value
	}
}

func listenerHasUDPOption(protocol string) bool {
	switch protocol {
	case "shadowsocks", "snell":
		return true
	default:
		return false
	}
}
