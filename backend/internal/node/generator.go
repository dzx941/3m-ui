package node

import (
	"encoding/json"
	"fmt"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"gopkg.in/yaml.v3"
)

// GenerateMihomoListeners converts database Node/Listener models to unified map arrays
func GenerateMihomoListeners(dbNodes []models.Listener) ([]map[string]interface{}, error) {
	var list []map[string]interface{}

	for _, dn := range dbNodes {
		if !dn.Enabled {
			continue
		}

		// Initialize core listener properties
		lm := map[string]interface{}{
			"name":   dn.Name,
			"type":   dn.Protocol,
			"port":   dn.Port,
			"listen": dn.BindAddress,
		}

		if dn.UDP {
			lm["udp"] = true
		}

		// Parse dynamic config attributes (passwords, certificates, uuid, flow, etc.)
		var configMap map[string]interface{}
		if dn.Config != "" {
			_ = json.Unmarshal([]byte(dn.Config), &configMap)
		}
		if configMap == nil {
			configMap = make(map[string]interface{})
		}

		// 1. TLS configuration block
		if dn.TLS {
			tlsMap := map[string]interface{}{
				"enable": true,
			}
			if cert, ok := configMap["certificate"]; ok {
				tlsMap["certificate"] = cert
			}
			if pk, ok := configMap["private_key"]; ok {
				tlsMap["private-key"] = pk
			}
			// Alternative tag
			if pk, ok := configMap["private-key"]; ok {
				tlsMap["private-key"] = pk
			}
			lm["tls"] = tlsMap
		}

		// 2. Protocols credential/users block
		proto := dn.Protocol
		var user map[string]interface{}

		switch proto {
		case "shadowsocks":
			if pwd, ok := configMap["password"]; ok {
				user = map[string]interface{}{
					"password": pwd,
				}
			}
		case "trojan":
			if pwd, ok := configMap["password"]; ok {
				user = map[string]interface{}{
					"password": pwd,
				}
			}
		case "vmess":
			if uid, ok := configMap["uuid"]; ok {
				user = map[string]interface{}{
					"uuid": uid,
				}
			}
		case "vless":
			userMap := map[string]interface{}{}
			if uid, ok := configMap["uuid"]; ok {
				userMap["uuid"] = uid
			}
			if flow, ok := configMap["flow"]; ok {
				userMap["flow"] = flow
			}
			if len(userMap) > 0 {
				user = userMap
			}
		case "hysteria2":
			if pwd, ok := configMap["password"]; ok {
				user = map[string]interface{}{
					"password": pwd,
				}
			}
		case "tuic":
			if pwd, ok := configMap["password"]; ok {
				user = map[string]interface{}{
					"password": pwd,
				}
			}
			if uuid, ok := configMap["uuid"]; ok {
				if user == nil {
					user = map[string]interface{}{}
				}
				user["uuid"] = uuid
			}
		}

		if user != nil {
			lm["users"] = []map[string]interface{}{user}
		}

		// Merge remaining extra parameters from dynamic config that are not already handled
		for k, v := range configMap {
			if k != "password" && k != "uuid" && k != "flow" && k != "certificate" && k != "private_key" && k != "private-key" {
				lm[k] = v
			}
		}

		list = append(list, lm)
	}

	return list, nil
}

// GenerateConfigYAML returns a full Mihomo listeners block serialized as YAML
func GenerateConfigYAML(dbNodes []models.Listener) (string, error) {
	listenersList, err := GenerateMihomoListeners(dbNodes)
	if err != nil {
		return "", err
	}

	root := map[string]interface{}{
		"listeners": listenersList,
	}

	bytes, err := yaml.Marshal(&root)
	if err != nil {
		return "", fmt.Errorf("failed to marshal listeners yaml: %w", err)
	}

	return string(bytes), nil
}
