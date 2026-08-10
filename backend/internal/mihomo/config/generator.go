package config

import (
	"encoding/json"
	"fmt"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"gorm.io/gorm"
	"gopkg.in/yaml.v3"
)

// ConfigEngine manages configuration generation, merging, and validations
type ConfigEngine struct {
	db *gorm.DB
}

func NewConfigEngine(db *gorm.DB) *ConfigEngine {
	return &ConfigEngine{
		db: db,
	}
}

// GenerateFinalConfig merges template, dynamic listeners, and custom fragments from db
func (ce *ConfigEngine) GenerateFinalConfig() (string, error) {
	// 1. Start with the default baseline template
	base := GetDefaultTemplate()

	// Convert base struct to a flexible map
	baseBytes, err := yaml.Marshal(base)
	if err != nil {
		return "", err
	}

	var merged map[string]interface{}
	if err := yaml.Unmarshal(baseBytes, &merged); err != nil {
		return "", err
	}

	// 2. Fetch and merge all enabled user-defined config fragments
	var customFragments []models.Config
	if err := ce.db.Where("enabled = ?", true).Find(&customFragments).Error; err == nil {
		for _, fragment := range customFragments {
			var fragMap map[string]interface{}
			// Try to parse YAML or JSON content
			if err := yaml.Unmarshal([]byte(fragment.Content), &fragMap); err == nil {
				for k, v := range fragMap {
					merged[k] = v
				}
			}
		}
	}

	// 3. Query active dynamic listeners and format into listeners block
	var dbListeners []models.Listener
	if err := ce.db.Where("enabled = ?", true).Find(&dbListeners).Error; err == nil && len(dbListeners) > 0 {
		var listenersList []map[string]interface{}
		for _, dl := range dbListeners {
			lm := map[string]interface{}{
				"name":   dl.Name,
				"type":   dl.Type,
				"listen": dl.Listen,
				"port":   dl.Port,
			}
			if dl.UDP {
				lm["udp"] = true
			}
			if dl.Proxy != "" {
				lm["proxy"] = dl.Proxy
			}
			if dl.Rule != "" {
				lm["rule"] = dl.Rule
			}
			if dl.Config != "" {
				var extra map[string]interface{}
				if err := json.Unmarshal([]byte(dl.Config), &extra); err == nil {
					for k, v := range extra {
						lm[k] = v
					}
				} else {
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
		merged["listeners"] = listenersList
	}

	// 4. Marshal final config to YAML
	finalBytes, err := yaml.Marshal(merged)
	if err != nil {
		return "", fmt.Errorf("failed to serialize merged configuration: %w", err)
	}

	return string(finalBytes), nil
}
