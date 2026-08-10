package listener

import (
	"encoding/json"
	"fmt"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"gopkg.in/yaml.v3"
)

// GenerateConfigYAML converts a list of GORM listeners to the Mihomo config YAML string
func GenerateConfigYAML(dbListeners []models.Listener) (string, error) {
	var listenersList []map[string]interface{}

	for _, dl := range dbListeners {
		// Only generate configuration for enabled listeners
		if !dl.Enabled {
			continue
		}

		// Initialize listener map with standard properties
		lm := map[string]interface{}{
			"name":   dl.Name,
			"type":   dl.Type,
			"listen": dl.Listen,
			"port":   dl.Port,
		}

		// Add optional standard fields
		if dl.UDP {
			lm["udp"] = true
		}
		if dl.Proxy != "" {
			lm["proxy"] = dl.Proxy
		}
		if dl.Rule != "" {
			lm["rule"] = dl.Rule
		}

		// Merge extra custom config parameters if present
		if dl.Config != "" {
			var extra map[string]interface{}
			// Try to parse as JSON first
			if err := json.Unmarshal([]byte(dl.Config), &extra); err == nil {
				for k, v := range extra {
					lm[k] = v
				}
			} else {
				// Fallback to YAML parse
				var extraYaml map[string]interface{}
				if err := yaml.Unmarshal([]byte(dl.Config), &extraYaml); err == nil {
					for k, v := range extraYaml {
						lm[k] = v
					}
				}
			}
		}

		listenersList = append(listenersList, lm)
	}

	// Create root structure
	cfg := MihomoConfig{
		Listeners: listenersList,
	}

	// Marshal to YAML string
	yamlBytes, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal mihomo config to yaml: %w", err)
	}

	return string(yamlBytes), nil
}
