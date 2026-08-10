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

	// Test GET /api/v1/mihomo/status (should be stopped initially)
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/mihomo/status", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["running"].(bool) != false {
			t.Fatalf("expected running to be false, got %v", resp)
		}
	}

	// Test POST /api/v1/mihomo/start (should start mock process)
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/mihomo/start", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["status"].(string) != "ok" {
			t.Fatalf("expected status 'ok', got %v", resp)
		}
	}

	// Test GET /api/v1/mihomo/status (should now be running)
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/mihomo/status", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["running"].(bool) != true {
			t.Fatalf("expected running to be true, got %v", resp)
		}
	}

	// Test POST /api/v1/mihomo/stop
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/mihomo/stop", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["status"].(string) != "ok" {
			t.Fatalf("expected status 'ok', got %v", resp)
		}
	}

	// Test GET /api/v1/mihomo/logs
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/mihomo/logs", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp []interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) == 0 {
			t.Fatal("expected non-empty logs array")
		}
	}

	// Test GET /api/v1/listeners (initially empty)
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/listeners", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp []interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected 0 listeners initially, got %d", len(resp))
		}
	}

	_ = os.RemoveAll("/tmp/3m-ui-router-test")
}
