package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Security SecurityConfig `yaml:"security"`
	Mihomo   MihomoConfig   `yaml:"mihomo"`
}

type ServerConfig struct {
	Port      int    `yaml:"port"`
	Mode      string `yaml:"mode"`
	PublicURL string `yaml:"public_url"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
}

type MihomoConfig struct {
	Binary string `yaml:"binary"`
	Config string `yaml:"config"`
}

var GlobalConfig *Config

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config YAML: %w", err)
	}

	GlobalConfig = &cfg
	return &cfg, nil
}

type SecurityConfig struct {
	CredentialKey string `yaml:"credential_key"`
}
