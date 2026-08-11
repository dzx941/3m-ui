package converter_test

import (
	"net/http"
	"testing"
	"time"

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

	err = db.AutoMigrate(
		&models.User{},
		&models.Listener{},
		&models.ListenerUser{},
		&models.AccessToken{},
		&models.ProxyUser{},
	)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func TestAccessTokenCRUDAndSubscriptions(t *testing.T) {
	db := setupTestDB(t)

	// 1. Create Access Token (user type)
	expire := time.Now().Add(10 * time.Minute)
	tokenObj := models.AccessToken{
		Name:     "Test Token 1",
		Token:    "test-user-token-123",
		Enabled:  true,
		ExpireAt: &expire,
		Type:     "user",
		TargetID: 1,
	}

	if err := db.Create(&tokenObj).Error; err != nil {
		t.Fatalf("failed to create access token: %v", err)
	}

	// 2. Query Access Token
	var fetched models.AccessToken
	if err := db.Where("token = ?", "test-user-token-123").First(&fetched).Error; err != nil {
		t.Fatalf("failed to find access token: %v", err)
	}

	if fetched.Name != "Test Token 1" {
		t.Errorf("expected token name 'Test Token 1', got '%s'", fetched.Name)
	}

	// 3. Disable Access Token
	fetched.Enabled = false
	if err := db.Save(&fetched).Error; err != nil {
		t.Fatalf("failed to disable token: %v", err)
	}

	var disabled models.AccessToken
	_ = db.First(&disabled, fetched.ID)
	if disabled.Enabled {
		t.Errorf("expected token to be disabled")
	}

	// 4. Token expiration check
	expiredTime := time.Now().Add(-1 * time.Second)
	tokenObj2 := models.AccessToken{
		Name:     "Expired Token",
		Token:    "expired-token-123",
		Enabled:  true,
		ExpireAt: &expiredTime,
		Type:     "user",
		TargetID: 1,
	}
	_ = db.Create(&tokenObj2)

	if tokenObj2.ExpireAt.Before(time.Now()) == false {
		t.Errorf("expected token to be expired")
	}

	// 5. Delete Access Token
	if err := db.Delete(&disabled).Error; err != nil {
		t.Fatalf("failed to delete token: %v", err)
	}

	var count int64
	db.Model(&models.AccessToken{}).Where("id = ?", disabled.ID).Count(&count)
	if count > 0 {
		t.Errorf("expected token to be deleted")
	}
}

func TestProxyUserSubscriptionGeneration(t *testing.T) {
	db := setupTestDB(t)

	// Create a Proxy User
	pu := models.ProxyUser{
		Username: "user-alice-credential",
		Enabled:  true,
	}
	_ = db.Create(&pu)

	// Create server listeners
	l1 := models.Listener{
		Name:        "hk-ss-listener",
		Protocol:    "shadowsocks",
		BindAddress: "0.0.0.0",
		Port:        10086,
		Enabled:     true,
		Config:      `{"password": "secret-alice-password", "cipher": "aes-256-gcm"}`,
	}
	_ = db.Create(&l1)

	// Bind listener to user
	lu := models.ListenerUser{
		ListenerID:  l1.ID,
		ProxyUserID: pu.ID,
	}
	_ = db.Create(&lu)

	// Create Access Token
	tokenObj := models.AccessToken{
		Name:     "Alice Subscription",
		Token:    "alice-token-xyz",
		Enabled:  true,
		Type:     "user",
		TargetID: pu.ID,
	}
	_ = db.Create(&tokenObj)

	// Generate raw config
	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/api/v1/client/sub/alice-token-xyz", nil)
	req.Host = "vps-public-domain.com:8080"

	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Server.Port = 8080

	data, err := converter.GenerateRawConfig(db, tokenObj, req)
	if err != nil {
		t.Fatalf("failed to generate raw config: %v", err)
	}

	// Parse generated YAML
	var generated map[string]interface{}
	if err := yaml.Unmarshal(data, &generated); err != nil {
		t.Fatalf("failed to unmarshal generated yaml: %v", err)
	}

	proxies, ok := generated["proxies"].([]interface{})
	if !ok || len(proxies) != 1 {
		t.Fatalf("expected 1 proxy in subscription, got: %v", generated["proxies"])
	}

	p := proxies[0].(map[string]interface{})
	if p["name"] != "hk-ss-listener" {
		t.Errorf("expected proxy name 'hk-ss-listener', got '%v'", p["name"])
	}
	if p["type"] != "shadowsocks" {
		t.Errorf("expected proxy type 'shadowsocks', got '%v'", p["type"])
	}
	if p["server"] != "vps-public-domain.com" {
		t.Errorf("expected resolved server address 'vps-public-domain.com', got '%v'", p["server"])
	}
	if p["port"] != 10086 {
		t.Errorf("expected port 10086, got '%v'", p["port"])
	}
}

func TestSubconverterUnreachableSafety(t *testing.T) {
	// If subconverter is down or unreachable, CallSubconverter should return a clean error instead of crashing
	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Server.Port = 8080

	// Set invalid URL to simulate unreachable subconverter
	converter.SubconverterURL = "http://127.0.0.1:9999"

	_, err := converter.CallSubconverter(config.GlobalConfig, "any-token", "singbox", []byte("proxies: []"))
	if err == nil {
		t.Fatalf("expected error from unreachable subconverter, got nil")
	}
}
