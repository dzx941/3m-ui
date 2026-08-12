package config

import "testing"

func TestIsMihomoListenerProtocol(t *testing.T) {
	supported := []string{
		"shadowsocks",
		"snell",
		"vmess",
		"vless",
		"trojan",
		"hysteria2",
		"hysteria2-realm",
		"tuic",
		"shadowquic",
		"anytls",
		"mieru",
		"sudoku",
		"trusttunnel",
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
		"wireguard",
		"",
	}
	for _, protocol := range rejected {
		if IsMihomoListenerProtocol(protocol) {
			t.Errorf("expected %q to be rejected", protocol)
		}
	}
}
