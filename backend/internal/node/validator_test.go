package node

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kazeyukiro/3m-ui/backend/internal/database/models"
)

func validNode(protocol, config string) *models.Listener {
	return &models.Listener{Name: "test", Protocol: protocol, Type: protocol, Port: "443", BindAddress: "0.0.0.0", Enabled: true, Config: config}
}

func TestValidateNodeRejectsUnknownField(t *testing.T) {
	err := ValidateNode(validNode("vless", `{"users":[{"username":"u","uuid":"00000000-0000-0000-0000-000000000001"}],"not-a-mihomo-field":true}`))
	if err == nil || !strings.Contains(err.Error(), "unsupported field") {
		t.Fatalf("expected unsupported field error, got %v", err)
	}
}
func TestValidateNodeRejectsTLSCertificateMismatch(t *testing.T) {
	err := ValidateNode(validNode("vless", `{"users":[{"username":"u","uuid":"00000000-0000-0000-0000-000000000001"}],"certificate":"/tmp/server.crt"}`))
	if err == nil || !strings.Contains(err.Error(), "certificate and private-key together") {
		t.Fatalf("expected certificate pair error, got %v", err)
	}
}
func TestValidateNodeRejectsRealityAndCertificateTogether(t *testing.T) {
	err := ValidateNode(validNode("vless", `{"users":[{"username":"u","uuid":"00000000-0000-0000-0000-000000000001"}],"certificate":"/tmp/server.crt","private-key":"/tmp/server.key","reality-config":{"dest":"example.com:443"}}`))
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive TLS modes") {
		t.Fatalf("expected TLS exclusivity error, got %v", err)
	}
}
func TestValidateNodeRejectsVmessFlow(t *testing.T) {
	err := ValidateNode(validNode("vmess", `{"users":[{"username":"u","uuid":"00000000-0000-0000-0000-000000000001","flow":"xtls-rprx-vision"}]}`))
	if err == nil || !strings.Contains(err.Error(), "vmess listener users[0] does not support flow") {
		t.Fatalf("expected VMess flow error, got %v", err)
	}
}
func TestValidateNodeRejectsVlessAlterID(t *testing.T) {
	err := ValidateNode(validNode("vless", `{"users":[{"username":"u","uuid":"00000000-0000-0000-0000-000000000001","alterId":0}]}`))
	if err == nil || !strings.Contains(err.Error(), "vless listener users[0] does not support alterId") {
		t.Fatalf("expected VLESS alterId error, got %v", err)
	}
}
func TestValidateNodeRejectsTuicV4V5Mix(t *testing.T) {
	err := ValidateNode(validNode("tuic", `{"users":{"00000000-0000-0000-0000-000000000001":"password"},"token":["TOKEN"]}`))
	if err == nil || !strings.Contains(err.Error(), "exactly one of users") {
		t.Fatalf("expected TUIC mode exclusivity error, got %v", err)
	}
}
func TestValidateNodeRejectsAnyTLSReality(t *testing.T) {
	err := ValidateNode(validNode("anytls", `{"users":{"u":"password"},"reality-config":{"dest":"example.com:443"}}`))
	if err == nil || !strings.Contains(err.Error(), "unsupported field") {
		t.Fatalf("expected AnyTLS schema rejection, got %v", err)
	}
}
func TestValidateNodeAcceptsVlessReality(t *testing.T) {
	err := ValidateNode(validNode("vless", `{"users":[{"username":"u","uuid":"00000000-0000-0000-0000-000000000001"}],"reality-config":{"dest":"example.com:443","private-key":"key","short-id":["0123456789abcdef"],"server-names":["example.com"]}}`))
	if err != nil {
		t.Fatalf("expected valid VLESS Reality config, got %v", err)
	}
}
func TestValidateNodeAcceptsTuicV5(t *testing.T) {
	err := ValidateNode(validNode("tuic", `{"users":{"00000000-0000-0000-0000-000000000001":"password"}}`))
	if err != nil {
		t.Fatalf("expected valid TUIC V5 config, got %v", err)
	}
}

func TestValidateNodeGeneratesVlessClientCredentialWhenMissing(t *testing.T) {
	listener := validNode("vless", `{}`)
	if err := ValidateNode(listener); err != nil {
		t.Fatalf("expected empty VLESS config to be completed, got %v", err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(listener.Config), &cfg); err != nil {
		t.Fatalf("generated config is invalid JSON: %v", err)
	}
	users, ok := cfg["users"].([]interface{})
	if !ok || len(users) != 1 {
		t.Fatalf("expected one generated VLESS user, got %#v", cfg["users"])
	}
	row, ok := users[0].(map[string]interface{})
	if !ok || row["username"] == "" || row["uuid"] == "" {
		t.Fatalf("generated VLESS credential is incomplete: %#v", users[0])
	}
}
