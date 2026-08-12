package listener_test

import (
	"os"
	"strings"
	"testing"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/listener"
)

func TestSupportedProtocols(t *testing.T) {
	expected := []string{
		"socks", "http", "tproxy", "redir", "mixed", "tunnel", "tun",
		"shadowsocks", "snell", "vmess", "vless", "trojan", "hysteria2",
		"hysteria2-realm", "tuic", "shadowquic", "anytls", "mieru", "sudoku", "trusttunnel",
	}
	got := listener.SupportedProtocols()
	if len(got) != len(expected) {
		t.Fatalf("expected %d supported protocols, got %d", len(expected), len(got))
	}
	for i, protocol := range expected {
		if got[i] != protocol {
			t.Fatalf("protocol %d: expected %q, got %q", i, protocol, got[i])
		}
	}
}

func TestGenerateConfigYAML(t *testing.T) {
	dbListeners := []models.Listener{
		{
			Name:    "test-mixed",
			Protocol: "mixed",
			Type:    "mixed",
			BindAddress: "0.0.0.0",
			Listen:  "127.0.0.1",
			Port:    1080,
			Enabled: true,
			UDP:     true,
		},
		{
			Name:    "test-socks",
			Type:    "socks",
			Listen:  "127.0.0.1",
			Port:    1081,
			Enabled: false,
		},
		{
			Name:    "test-custom",
			Type:    "shadowsocks",
			Listen:  "0.0.0.0",
			Port:    1082,
			Enabled: true,
			Config:  `{"cipher":"aes-256-gcm","password":"pass"}`,
		},
	}

	yamlStr, err := listener.GenerateConfigYAML(dbListeners)
	if err != nil {
		t.Fatalf("failed to generate yaml: %v", err)
	}
	if !strings.Contains(yamlStr, "test-mixed") || !strings.Contains(yamlStr, "listen: 0.0.0.0") {
		t.Fatal("expected BindAddress to be used as Mihomo listen address")
	}
	if strings.Contains(yamlStr, "test-socks") {
		t.Fatal("expected disabled listener to be ignored")
	}
	if !strings.Contains(yamlStr, "cipher: aes-256-gcm") {
		t.Fatal("expected protocol-specific config to be preserved")
	}
}

func TestGenerateConfigYAMLRejectsUnknownProtocol(t *testing.T) {
	_, err := listener.GenerateConfigYAML([]models.Listener{{
		Name: "invalid",
		Type: "not-a-mihomo-listener",
		Port: 1000,
		Enabled: true,
	}})
	if err == nil {
		t.Fatal("expected unsupported listener protocol to be rejected")
	}
}

func TestListenerServiceLifecycle(t *testing.T) {
	_ = os.RemoveAll("/tmp/3m-ui-listener-service-test")
}
