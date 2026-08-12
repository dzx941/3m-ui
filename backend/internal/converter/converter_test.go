package converter_test

import (
	"net/http"
	"testing"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/converter"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	if err = db.AutoMigrate(&models.User{}, &models.Listener{}, &models.AccessToken{}); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	return db
}

func TestAccessTokenCRUDAndListenerBinding(t *testing.T) {
	db := setupTestDB(t)

	listener := models.Listener{
		Name:        "test-vless",
		Protocol:    "vless",
		BindAddress: "0.0.0.0",
		Port:        10086,
		Enabled:     true,
		Config:      `{"uuid":"11111111-1111-1111-1111-111111111111","flow":"xtls-rprx-vision"}`,
	}
	if err := db.Create(&listener).Error; err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	token := models.AccessToken{
		Name:       "Test Token",
		Token:      "test-listener-token",
		Enabled:    true,
		ListenerID: listener.ID,
	}
	if err := db.Create(&token).Error; err != nil {
		t.Fatalf("failed to create access token: %v", err)
	}

	var fetched models.AccessToken
	if err := db.Where("token = ?", token.Token).First(&fetched).Error; err != nil {
		t.Fatalf("failed to find access token: %v", err)
	}
	if fetched.ListenerID != listener.ID {
		t.Fatalf("expected listener id %d, got %d", listener.ID, fetched.ListenerID)
	}

	fetched.Enabled = false
	if err := db.Save(&fetched).Error; err != nil {
		t.Fatalf("failed to disable token: %v", err)
	}

	var disabled models.AccessToken
	if err := db.First(&disabled, fetched.ID).Error; err != nil {
		t.Fatalf("failed to fetch disabled token: %v", err)
	}
	if disabled.Enabled {
		t.Errorf("expected token to be disabled")
	}

	if err := db.Delete(&disabled).Error; err != nil {
		t.Fatalf("failed to delete access token: %v", err)
	}
}

func TestListenerSubscriptionGeneration(t *testing.T) {
	db := setupTestDB(t)

	listener := models.Listener{
		Name:        "hk-vless-listener",
		Protocol:    "vless",
		BindAddress: "0.0.0.0",
		Port:        10086,
		Enabled:     true,
		UDP:         true,
		Config:      `{"uuid":"11111111-1111-1111-1111-111111111111","flow":"xtls-rprx-vision","network":"tcp"}`,
	}
	if err := db.Create(&listener).Error; err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	token := models.AccessToken{
		Name:       "Listener Subscription",
		Token:      "listener-token-xyz",
		Enabled:    true,
		ListenerID: listener.ID,
	}
	if err := db.Create(&token).Error; err != nil {
		t.Fatalf("failed to create access token: %v", err)
	}

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/api/v1/client/sub/listener-token-xyz", nil)
	req.Host = "vps-public-domain.com:8080"
	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Server.Port = 8080

	data, err := converter.GenerateRawConfig(db, token, req)
	if err != nil {
		t.Fatalf("failed to generate raw config: %v", err)
	}

	var generated map[string]interface{}
	if err := yaml.Unmarshal(data, &generated); err != nil {
		t.Fatalf("failed to unmarshal generated yaml: %v", err)
	}

	proxies, ok := generated["proxies"].([]interface{})
	if !ok || len(proxies) != 1 {
		t.Fatalf("expected 1 proxy in subscription, got: %v", generated["proxies"])
	}
	p := proxies[0].(map[string]interface{})
	if p["name"] != "hk-vless-listener" {
		t.Errorf("expected proxy name 'hk-vless-listener', got '%v'", p["name"])
	}
	if p["type"] != "vless" {
		t.Errorf("expected proxy type 'vless', got '%v'", p["type"])
	}
	if p["server"] != "vps-public-domain.com" {
		t.Errorf("expected resolved server address 'vps-public-domain.com', got '%v'", p["server"])
	}
	if p["port"] != 10086 {
		t.Errorf("expected port 10086, got '%v'", p["port"])
	}
	if p["uuid"] != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected UUID to be preserved, got '%v'", p["uuid"])
	}
	if p["flow"] != "xtls-rprx-vision" {
		t.Errorf("expected flow to be preserved, got '%v'", p["flow"])
	}
	if p["udp"] != true {
		t.Errorf("expected UDP to be preserved")
	}
}

func TestSubconverterUnreachableSafety(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Server.Port = 8080
	converter.SubconverterURL = "http://127.0.0.1:9999"
	if _, err := converter.CallSubconverter(config.GlobalConfig, "any-token", "singbox", []byte("proxies: []")); err == nil {
		t.Fatalf("expected error from unreachable subconverter, got nil")
	}
}
