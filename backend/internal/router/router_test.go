package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/auth"
	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/listener"
	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
	"github.com/dzx941/3m-ui/backend/internal/router"
	"github.com/dzx941/3m-ui/backend/internal/system"
	"github.com/dzx941/3m-ui/backend/internal/traffic"
)

func TestHealthAndMihomoAPIs(t *testing.T) {
	_ = os.RemoveAll("/tmp/3m-ui-router-test")

	cfg := &config.Config{}
	cfg.Server.Mode = "debug"
	cfg.Database.Path = "/tmp/3m-ui-router-test/db.sqlite"
	cfg.Mihomo.Binary = "/tmp/dummy-nonexistent"
	cfg.Mihomo.Config = "/tmp/3m-ui-router-test/config.yaml"
	cfg.JWT.Secret = "super-secret-token-key-for-testing-purposes"

	db, err := database.InitDB(cfg.Database.Path)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	mihomo.InitService(cfg)
	listener.InitService(db, cfg.Mihomo.Config)
	system.InitService()

	deps := router.Dependencies{
		DB:           db,
		Config:       cfg,
		Mihomo:       mihomo.GlobalService,
		Listener:     listener.GlobalService,
		Traffic:      traffic.NewService(),
		ConfigEngine: mihomoConfig.NewConfigEngine(db),
	}
	r := router.SetupRouter(cfg, deps)

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
			t.Fatalf("expected health status 'ok', got '%v'", resp["status"])
		}
	}

	{
		token, _, err := auth.GenerateToken(cfg.JWT.Secret, 1, "admin", "admin", 1*time.Hour)
		if err != nil {
			t.Fatalf("failed to generate testing JWT: %v", err)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+token)
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
