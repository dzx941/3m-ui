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
