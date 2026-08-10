package config

// GetDefaultTemplate returns a default base config block
func GetDefaultTemplate() *MihomoConfig {
	return &MihomoConfig{
		Mode:               "rule",
		LogLevel:           "info",
		AllowLan:           true,
		IPv6:               false,
		ExternalController: "127.0.0.1:9090",
		Secret:             "3m-ui-default-secret-key",
		DNS: map[string]interface{}{
			"enable":        true,
			"listen":        "0.0.0.0:1053",
			"enhanced-mode": "fake-ip",
			"nameserver": []string{
				"119.29.29.29",
				"223.5.5.5",
			},
		},
		Proxies:     []map[string]interface{}{},
		ProxyGroups: []map[string]interface{}{},
		Rules: []string{
			"GEOIP,CN,DIRECT",
			"MATCH,DIRECT",
		},
	}
}
