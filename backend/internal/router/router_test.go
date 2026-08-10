package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/listener"
	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	"github.com/dzx941/3m-ui/backend/internal/router"
	"github.com/dzx941/3m-ui/backend/internal/system"
)

func TestHealthAndMihomoAPIs(t *testing.T) {
	_ = os.RemoveAll("/tmp/3m-ui-router-test")

	cfg := &config.Config{}
	cfg.Server.Mode = "debug"
	cfg.Database.Path = "/tmp/3m-ui-router-test/db.sqlite"
	cfg.Mihomo.Binary = "/tmp/dummy-nonexistent"
	cfg.Mihomo.Config = "/tmp/3m-ui-router-test/config.yaml"

	// Init DB
	db, err := database.InitDB(cfg.Database.Path)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	// Init service layer
	mihomo.InitService(cfg)
	listener.InitService(db, cfg.Mihomo.Config)
	system.InitService()

	r := router.SetupRouter(cfg)

	// Test GET /api/v1/health
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["status"] != "ok" {
			t.Fatalf("expected health status 'ok', got '%v'", resp)
		}
	}

	// Test GET /api/v1/dashboard (Unified Aggregator)
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if _, exists := resp["mihomo"]; !exists {
			t.Fatal("expected aggregator to contain 'mihomo'")
		}
		if _, exists := resp["system"]; !exists {
			t.Fatal("expected aggregator to contain 'system'")
		}
		if _, exists := resp["listeners"]; !exists {
			t.Fatal("expected aggregator to contain 'listeners'")
		}
	}

	_ = os.RemoveAll("/tmp/3m-ui-router-test")
}
