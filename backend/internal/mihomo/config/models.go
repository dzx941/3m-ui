package config

// MihomoConfig represents the full structure of Mihomo configuration
type MihomoConfig struct {
	Mode               string                   `yaml:"mode,omitempty"`
	Port               int                      `yaml:"port,omitempty"`
	SocksPort          int                      `yaml:"socks-port,omitempty"`
	MixedPort          int                      `yaml:"mixed-port,omitempty"`
	AllowLan           bool                     `yaml:"allow-lan,omitempty"`
	LogLevel           string                   `yaml:"log-level,omitempty"`
	IPv6               bool                     `yaml:"ipv6,omitempty"`
	ExternalController string                   `yaml:"external-controller,omitempty"`
	Secret             string                   `yaml:"secret,omitempty"`
	DNS                map[string]interface{}   `yaml:"dns,omitempty"`
	Listeners          []map[string]interface{} `yaml:"listeners,omitempty"`
	Proxies            []map[string]interface{} `yaml:"proxies,omitempty"`
	ProxyGroups        []map[string]interface{} `yaml:"proxy-groups,omitempty"`
	Rules              []string                 `yaml:"rules,omitempty"`
}
