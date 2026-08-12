package config

import "testing"

func TestIsMihomoListenerProtocol(t *testing.T) {
	supported := []string{
		"shadowsocks",
		"vmess",
		"vless",
		"trojan",
		"hysteria2",
		"tuic",
		"shadowquic",
		"anytls",
		"mieru",
	}
	for _, protocol := range supported {
		if !IsMihomoListenerProtocol(protocol) {
			t.Errorf("expected %q to be supported", protocol)
		}
	}
}

func TestIsMihomoListenerProtocolRejectsInboundOnlyTypes(t *testing.T) {
	rejected := []string{
		"socks",
		"http",
		"tproxy",
		"redir",
		"mixed",
		"tunnel",
		"tun",
		"hysteria2-realm",
		"",
	}
	for _, protocol := range rejected {
		if IsMihomoListenerProtocol(protocol) {
			t.Errorf("expected %q to be rejected", protocol)
		}
	}
}
