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

// IsMihomoListenerProtocol reports whether protocol is a listener protocol
// that 3m-ui currently knows how to export as a client proxy configuration.
//
// These are intentionally different from Mihomo's general inbound/listener
// surface. SOCKS, HTTP, TPROXY, REDIR, Mixed and Tunnel are local inbound
// transport/listener types and are not exported as remote client proxy nodes
// by 3m-ui. TUN is also not a protocol and is therefore excluded here.
// Auxiliary listener services such as hysteria2-realm are excluded as well.
func IsMihomoListenerProtocol(protocol string) bool {
	switch protocol {
	case "shadowsocks",
		"vmess",
		"vless",
		"trojan",
		"hysteria2",
		"tuic",
		"shadowquic",
		"anytls",
		"mieru":
		return true
	default:
		return false
	}
}

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
