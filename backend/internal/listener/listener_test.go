package listener_test

import (
	"os"
	"strings"
	"testing"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/listener"
)

func TestGenerateConfigYAML(t *testing.T) {
	dbListeners := []models.Listener{
		{
			Name:    "test-mixed",
			Type:    "mixed",
			Listen:  "0.0.0.0",
			Port:    1080,
			Enabled: true,
			UDP:     true,
		},
		{
			Name:    "test-socks",
			Type:    "socks",
			Listen:  "127.0.0.1",
			Port:    1081,
			Enabled: false, // Should be ignored
		},
		{
			Name:    "test-custom",
			Type:    "shadowsocks",
			Listen:  "0.0.0.0",
			Port:    1082,
			Enabled: true,
			Config:  `{"cipher": "aes-256-gcm", "password": "pass"}`,
		},
	}

	yamlStr, err := listener.GenerateConfigYAML(dbListeners)
	if err != nil {
		t.Fatalf("failed to generate yaml: %v", err)
	}

	if !strings.Contains(yamlStr, "test-mixed") {
		t.Fatal("expected yaml to contain test-mixed listener name")
	}

	if strings.Contains(yamlStr, "test-socks") {
		t.Fatal("expected yaml NOT to contain test-socks (since enabled=false)")
	}

	if !strings.Contains(yamlStr, "cipher: aes-256-gcm") {
		t.Fatal("expected yaml to contain custom config properties")
	}
}

func TestListenerServiceLifecycle(t *testing.T) {
	_ = os.RemoveAll("/tmp/3m-ui-listener-service-test")
}
