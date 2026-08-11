package mihomo

type ConfigManager struct {
	configPath string
}

func NewConfigManager(path string) *ConfigManager {
	return &ConfigManager{
		configPath: path,
	}
}

// SaveConfig saves the configuration YAML file
func (cm *ConfigManager) SaveConfig(content string) error {
	// For future implementation
	return nil
}
