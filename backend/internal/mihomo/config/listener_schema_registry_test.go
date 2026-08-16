package config

import "testing"

func TestMihomoListenerSchemaRegistry(t *testing.T) {
	for _, protocol := range MihomoListenerProtocols {
		schema, ok := GetMihomoListenerSchema(protocol)
		if !ok {
			t.Fatalf("protocol %q has no schema", protocol)
		}
		if schema.Protocol != protocol {
			t.Fatalf("schema protocol mismatch: got %q want %q", schema.Protocol, protocol)
		}
		if len(schema.Fields) == 0 {
			t.Fatalf("protocol %q has no registered fields", protocol)
		}
	}
}

func TestMihomoListenerSchemaRejectsNonListenerProtocols(t *testing.T) {
	for _, protocol := range []string{"socks", "http", "mixed", "redir", "tproxy", "tun", "tunnel"} {
		if _, ok := GetMihomoListenerSchema(protocol); ok {
			t.Fatalf("non-distributable protocol %q must not have a node schema", protocol)
		}
	}
}

func TestMihomoListenerSchemaIncludesFieldsUsedByForms(t *testing.T) {
	cases := map[string][]string{
		"shadowsocks": {"cipher", "password", "simple-obfs", "shadow-tls", "res-tls", "jls-config", "kcp-tun", "mux-option"},
		"snell":       {"psk", "version", "obfs-opts", "shadow-tls", "res-tls", "jls-config"},
		"vless":       {"users", "ws-path", "grpc-service-name", "xhttp-config", "reality-config"},
		"vmess":       {"users", "ws-path", "grpc-service-name", "mekya-config", "mkcp-config", "reality-config"},
		"trojan":      {"users", "ws-path", "grpc-service-name", "reality-config", "ss-option"},
		"hysteria2":   {"users", "obfs", "certificate", "private-key", "alpn"},
		"tuic":        {"users", "token", "certificate", "private-key", "congestion-controller"},
		"anytls":      {"users", "certificate", "private-key", "padding-scheme"},
	}
	for protocol, fields := range cases {
		schema, ok := GetMihomoListenerSchema(protocol)
		if !ok {
			t.Fatalf("protocol %q has no schema", protocol)
		}
		for _, field := range fields {
			if _, ok := schema.Fields[field]; !ok {
				t.Errorf("protocol %q is missing form field %q", protocol, field)
			}
		}
	}
}
