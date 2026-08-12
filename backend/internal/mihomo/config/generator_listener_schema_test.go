package config

import (
	"testing"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/user"
)

func TestGenerateListenersUsesNativeSchema(t *testing.T) {
	listeners := []models.Listener{
		{Name: "ss", Protocol: "shadowsocks", Port: 443, BindAddress: "0.0.0.0", Enabled: true, Config: `{"cipher":"aes-256-gcm"}`},
		{Name: "vless", Protocol: "vless", Port: 8443, BindAddress: "0.0.0.0", Enabled: true, Config: `{"flow":"xtls-rprx-vision","certificate":"cert","private-key":"key","reality-config":{"private-key":"secret"}}`},
	}
	creds := map[uint][]user.Credential{
		1: {{Username: "alice", Password: "ss-pass"}},
		2: {{Username: "alice", UUID: "11111111-1111-4111-8111-111111111111"}},
	}
	listeners[0].ID = 1
	listeners[1].ID = 2

	result, err := generateListeners(listeners, creds)
	if err != nil {
		t.Fatalf("generateListeners failed: %v", err)
	}
	if result[0]["password"] != "ss-pass" {
		t.Fatal("Shadowsocks listener password was not generated")
	}
	vlessUsers, ok := result[1]["users"].([]map[string]interface{})
	if !ok || len(vlessUsers) != 1 || vlessUsers[0]["uuid"] == "" {
		t.Fatal("VLESS listener users were not generated")
	}
	if result[1]["tls"] != nil {
		t.Fatal("listener TLS must not use a nested tls object")
	}
	if result[1]["certificate"] != "cert" || result[1]["private-key"] != "key" {
		t.Fatal("listener certificate/private-key fields were not preserved")
	}
}

func TestGenerateListenersRejectsExcludedProtocols(t *testing.T) {
	for _, protocol := range []string{"socks", "http", "tproxy", "redir", "mixed", "tunnel", "tun", "wireguard"} {
		_, err := generateListeners([]models.Listener{{Name: "bad", Protocol: protocol, Port: 1080, Enabled: true}}, nil)
		if err == nil {
			t.Fatalf("expected protocol %q to be rejected", protocol)
		}
	}
}
