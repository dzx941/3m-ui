package config

// GetDefaultTemplate returns a deliberately minimal, localhost-safe base
// configuration. Listener definitions are appended from the database.
func GetDefaultTemplate() *MihomoConfig {
	return &MihomoConfig{
		Mode:               "rule",
		LogLevel:           "info",
		AllowLan:           false,
		IPv6:               false,
		ExternalController: "127.0.0.1:9090",
		// The controller is bound to loopback, so an authentication secret is
		// not required by default. Never ship a hard-coded reusable secret.
		Secret: "",
		DNS: map[string]interface{}{
			"enable": false,
		},
		Proxies:     []map[string]interface{}{},
		ProxyGroups: []map[string]interface{}{},
		Rules: []string{
			"GEOIP,CN,DIRECT",
			"MATCH,DIRECT",
		},
	}
}
