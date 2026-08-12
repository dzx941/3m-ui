package config

import (
	"testing"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/user"
)

func TestGenerateListenersProtocolShapes(t *testing.T) {
	listeners := []models.Listener{
		{Name: "ss", Protocol: "shadowsocks", Port: 10001, BindAddress: "0.0.0.0", Enabled: true, Config: `{"cipher":"2022-blake3-aes-256-gcm"}`},
		{Name: "vmess", Protocol: "vmess", Port: 10002, BindAddress: "0.0.0.0", Enabled: true, Config: `{"alterId":0}`},
		{Name: "vless", Protocol: "vless", Port: 10003, BindAddress: "0.0.0.0", Enabled: true, Config: `{"flow":"xtls-rprx-vision"}`},
		{Name: "trojan", Protocol: "trojan", Port: 10004, BindAddress: "0.0.0.0", Enabled: true, Config: `{"certificate":"server.crt","private-key":"server.key"}`},
		{Name: "hy2", Protocol: "hysteria2", Port: 10005, BindAddress: "0.0.0.0", Enabled: true, Config: `{"up":1000,"down":1000,"certificate":"server.crt","private-key":"server.key"}`},
		{Name: "tuic", Protocol: "tuic", Port: 10006, BindAddress: "0.0.0.0", Enabled: true, Config: `{"congestion-controller":"bbr","certificate":"server.crt","private-key":"server.key"}`},
	}
	for i := range listeners { listeners[i].ID = uint(i + 1) }

	creds := map[uint][]user.Credential{
		1: {{Username: "alice", Password: "ss-pass"}},
		2: {{Username: "alice", UUID: "11111111-1111-4111-8111-111111111111"}},
		3: {{Username: "alice", UUID: "22222222-2222-4222-8222-222222222222"}},
		4: {{Username: "alice", Password: "trojan-pass"}},
		5: {{Username: "alice", Password: "hy-pass"}},
		6: {{Username: "alice", UUID: "33333333-3333-4333-8333-333333333333", Password: "tuic-pass"}},
	}

	result, err := generateListeners(listeners, creds)
	if err != nil { t.Fatalf("generateListeners failed: %v", err) }

	if got := result[0]["password"]; got != "ss-pass" { t.Fatalf("shadowsocks password = %v", got) }
	if _, ok := result[0]["users"]; ok { t.Fatal("shadowsocks listener must not contain a users list") }

	vmessUsers := result[1]["users"].([]map[string]interface{})
	if vmessUsers[0]["uuid"] != "11111111-1111-4111-8111-111111111111" { t.Fatal("vmess uuid was not generated") }

	vlessUsers := result[2]["users"].([]map[string]interface{})
	if vlessUsers[0]["flow"] != "xtls-rprx-vision" { t.Fatal("vless flow must be attached to the user") }

	trojanUsers := result[3]["users"].([]map[string]interface{})
	if trojanUsers[0]["username"] != "alice" || trojanUsers[0]["password"] != "trojan-pass" { t.Fatal("trojan user shape is invalid") }
	if result[3]["tls"] != nil { t.Fatal("listener TLS must not be represented as a nested tls object") }
	if result[3]["certificate"] != "server.crt" || result[3]["private-key"] != "server.key" { t.Fatal("trojan certificate fields are invalid") }

	hyUsers := result[4]["users"].(map[string]string)
	if hyUsers["alice"] != "hy-pass" { t.Fatal("hysteria2 users must be a username/password map") }

	tuicUsers := result[5]["users"].(map[string]string)
	if tuicUsers["33333333-3333-4333-8333-333333333333"] != "tuic-pass" { t.Fatal("tuic users must be a uuid/password map") }
}

func TestGenerateListenersRejectsMultipleShadowsocksUsers(t *testing.T) {
	listeners := []models.Listener{{ID: 1, Name: "ss", Protocol: "shadowsocks", Port: 10001, BindAddress: "0.0.0.0", Enabled: true, Config: `{"cipher":"aes-256-gcm"}`}}
	creds := map[uint][]user.Credential{1: {{Username: "a", Password: "one"}, {Username: "b", Password: "two"}}}
	if _, err := generateListeners(listeners, creds); err == nil { t.Fatal("expected multiple Shadowsocks credentials to be rejected") }
}
