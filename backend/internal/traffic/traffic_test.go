package traffic_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"github.com/dzx941/3m-ui/backend/internal/node"
	"github.com/dzx941/3m-ui/backend/internal/security"
	"github.com/dzx941/3m-ui/backend/internal/traffic"
	"github.com/dzx941/3m-ui/backend/internal/user"
)

func TestTrafficCollectionAndMapping(t *testing.T) {
	security.InitCredentialKey("test-secret")
	dbPath := t.TempDir() + "/test.db"
	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize services
	user.InitService(db)
	node.InitService(db, "/tmp/mihomo.yaml")
	traffic.InitService()

	// 1. Create a dummy Listener and ProxyUser
	l := &models.Listener{
		Name:     "hk-vless",
		Protocol: "vless",
		Port:     10086,
		Enabled:  true,
	}
	if err := db.Create(l).Error; err != nil {
		t.Fatal(err)
	}

	u, err := user.GlobalService.Create(user.CreateInput{
		Username: "alice",
		UUID:     "9f3d64c0-ff0c-4da6-928e-57b1a030be90",
	})
	if err != nil {
		t.Fatal(err)
	}

	// 2. Set up a Mock Mihomo server to handle /traffic and /connections
	mockMihomo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/traffic" {
			w.Write([]byte(`{"up": 500, "down": 1000}`))
			return
		}
		if r.URL.Path == "/connections" {
			w.Write([]byte(`{
				"downloadTotal": 50000,
				"uploadTotal": 20000,
				"connections": [
					{
						"id": "conn1",
						"network": "tcp",
						"upload": 1000,
						"download": 2000,
						"metadata": {
							"network": "tcp",
							"type": "vless",
							"sourceIP": "127.0.0.1",
							"sourcePort": "54321",
							"destinationIP": "1.1.1.1",
							"destinationPort": "443",
							"host": "one.one.one.one",
							"inboundName": "hk-vless",
							"inboundUser": "alice"
						}
					}
				]
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockMihomo.Close()

	// 3. Write dummy config to path to simulate the configured external controller
	configContent := "external-controller: " + mockMihomo.URL[7:] + "\nsecret: test\n"
	configPath := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 4. Run collector
	collector := traffic.NewCollector(configPath, traffic.GlobalService)
	if err := collector.Collect(context.Background()); err != nil {
		t.Fatal(err)
	}

	// Verify database changes and online status
	var updatedUser models.ProxyUser
	if err := db.First(&updatedUser, u.ID).Error; err != nil {
		t.Fatal(err)
	}

	if !updatedUser.Online {
		t.Errorf("expected user to be online")
	}

	if updatedUser.UploadBytes != 1000 || updatedUser.DownloadBytes != 2000 {
		t.Errorf("expected 1000 up, 2000 down, got %d up, %d down", updatedUser.UploadBytes, updatedUser.DownloadBytes)
	}

	// Verify TrafficRecord was created
	var records []models.TrafficRecord
	if err := db.Find(&records).Error; err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Errorf("expected 1 TrafficRecord, got %d", len(records))
	} else {
		if records[0].UploadBytes != 1000 || records[0].DownloadBytes != 2000 {
			t.Errorf("expected 1000 up / 2000 down in TrafficRecord, got %d / %d", records[0].UploadBytes, records[0].DownloadBytes)
		}
	}

	// Run collector a second time with incremental changes
	// We update the mock server behavior or simulate it. We can just keep the connection active with higher values.
	mockMihomo.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/traffic" {
			w.Write([]byte(`{"up": 600, "down": 1200}`))
			return
		}
		if r.URL.Path == "/connections" {
			w.Write([]byte(`{
				"downloadTotal": 60000,
				"uploadTotal": 25000,
				"connections": [
					{
						"id": "conn1",
						"network": "tcp",
						"upload": 1500,
						"download": 2800,
						"metadata": {
							"network": "tcp",
							"type": "vless",
							"sourceIP": "127.0.0.1",
							"sourcePort": "54321",
							"destinationIP": "1.1.1.1",
							"destinationPort": "443",
							"host": "one.one.one.one",
							"inboundName": "hk-vless",
							"inboundUser": "alice"
						}
					}
				]
			}`))
			return
		}
	})

	if err := collector.Collect(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := db.First(&updatedUser, u.ID).Error; err != nil {
		t.Fatal(err)
	}

	// Incremental: diffUp = 1500 - 1000 = 500, diffDown = 2800 - 2000 = 800
	// Total upload_bytes = 1000 + 500 = 1500, download_bytes = 2000 + 800 = 2800
	if updatedUser.UploadBytes != 1500 || updatedUser.DownloadBytes != 2800 {
		t.Errorf("expected 1500 up, 2800 down after incremental update, got %d up, %d down", updatedUser.UploadBytes, updatedUser.DownloadBytes)
	}

	// Run collector a third time with no connections (alice goes offline)
	mockMihomo.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/traffic" {
			w.Write([]byte(`{"up": 0, "down": 0}`))
			return
		}
		if r.URL.Path == "/connections" {
			w.Write([]byte(`{
				"downloadTotal": 60000,
				"uploadTotal": 25000,
				"connections": []
			}`))
			return
		}
	})

	if err := collector.Collect(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := db.First(&updatedUser, u.ID).Error; err != nil {
		t.Fatal(err)
	}

	if updatedUser.Online {
		t.Errorf("expected user to be offline")
	}
}
