package config

import (
	"encoding/json"
	"fmt"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// ConfigEngine manages configuration generation, merging, and validation.
type ConfigEngine struct {
	db *gorm.DB
}

func NewConfigEngine(db *gorm.DB) *ConfigEngine { return &ConfigEngine{db: db} }

// GenerateFinalConfig merges the standard template, enabled custom fragments,
// and the current server nodes/users into one complete Mihomo configuration.
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
		return "", fmt.Errorf("load enabled nodes: %w", err)
	}

	credentials := make(map[uint][]user.Credential)
	if user.GlobalService != nil {
		credentials, err = user.GlobalService.ActiveCredentialsByListener()
		if err != nil {
			return "", fmt.Errorf("load proxy user credentials: %w", err)
		}
	}

	merged["listeners"] = generateListeners(listeners, credentials)
	finalBytes, err := yaml.Marshal(merged)
	if err != nil {
		return "", fmt.Errorf("serialize final configuration: %w", err)
	}
	return string(finalBytes), nil
}

type listenerCredential struct {
	Password string
	UUID     string
}

func generateListeners(listeners []models.Listener, creds map[uint][]user.Credential) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(listeners))
	for _, l := range listeners {
		m := map[string]interface{}{
			"name": l.Name, "type": l.Protocol, "port": l.Port, "listen": l.BindAddress,
			"users": []map[string]interface{}{},
		}
		if l.UDP {
			m["udp"] = true
		}
		var configMap map[string]interface{}
		if l.Config != "" {
			_ = json.Unmarshal([]byte(l.Config), &configMap)
		}
		if configMap == nil {
			configMap = map[string]interface{}{}
		}

		if l.TLS {
			tls := map[string]interface{}{"enable": true}
			if v, ok := configMap["certificate"]; ok {
				tls["certificate"] = v
			}
			if v, ok := configMap["private_key"]; ok {
				tls["private-key"] = v
			}
			if v, ok := configMap["private-key"]; ok {
				tls["private-key"] = v
			}
			m["tls"] = tls
		}

		for k, v := range configMap {
			switch k {
			case "password", "uuid", "flow", "certificate", "private_key", "private-key":
				continue
			default:
				m[k] = v
			}
		}

		for _, cred := range creds[l.ID] {
			u := map[string]interface{}{}
			switch l.Protocol {
			case "shadowsocks", "trojan", "hysteria2":
				if cred.Password != "" {
					u["password"] = cred.Password
				}
			case "vmess":
				if cred.UUID != "" {
					u["uuid"] = cred.UUID
				}
			case "vless":
				if cred.UUID != "" {
					u["uuid"] = cred.UUID
				}
				if flow, ok := configMap["flow"]; ok {
					u["flow"] = flow
				}
			case "tuic":
				if cred.UUID != "" {
					u["uuid"] = cred.UUID
				}
			}
			if len(u) > 0 {
				m["users"] = append(m["users"].([]map[string]interface{}), u)
			}
		}
		result = append(result, m)
	}
	return result
}
