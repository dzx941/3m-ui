package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kazeyukiro/3m-ui/backend/internal/auth"
	"github.com/kazeyukiro/3m-ui/backend/internal/config"
	"github.com/kazeyukiro/3m-ui/backend/internal/database"
	"github.com/kazeyukiro/3m-ui/backend/internal/mihomo"
	"github.com/kazeyukiro/3m-ui/backend/internal/router"
	"github.com/kazeyukiro/3m-ui/backend/internal/system"
	"github.com/kazeyukiro/3m-ui/backend/internal/traffic"
)

func TestHealthAndMihomoAPIs(t *testing.T) {
	_ = os.RemoveAll("/tmp/3m-ui-router-test")

	cfg := &config.Config{}
	cfg.Server.Mode = "debug"
	cfg.Database.Path = "/tmp/3m-ui-router-test/db.sqlite"
	cfg.Mihomo.Binary = "/tmp/dummy-nonexistent"
	cfg.Mihomo.Config = "/tmp/3m-ui-router-test/config.yaml"
	cfg.JWT.Secret = "super-secret-token-key-for-testing-purposes"
	cfg.Security.CORSOrigins = nil

	db, err := database.InitDB(cfg.Database.Path)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	mihomoSvc := mihomo.NewService(cfg)
	systemSvc := system.NewService()
	trafficSvc := traffic.NewService()

	r := router.SetupRouterWithDeps(router.Deps{
		DB:      db,
		Config:  cfg,
		Mihomo:  mihomoSvc,
		System:  systemSvc,
		Traffic: trafficSvc,
	})

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
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Fatalf("expected CORS allow-origin *, got %q", got)
		}
	}

	{
		if _, _, _, err := auth.EnsureAdmin(db, cfg.Database.Path); err != nil {
			t.Fatalf("ensure admin: %v", err)
		}
		result, err := auth.Login(db, cfg.JWT.Secret, auth.LoginInput{Username: "admin", Password: "admin"})
		if err != nil {
			t.Fatalf("login: %v", err)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+result.Token)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK && w.Code != http.StatusForbidden {
			t.Fatalf("expected status 200 or 403, got %d body=%s", w.Code, w.Body.String())
		}

		if w.Code == http.StatusOK {
			var resp map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			for _, key := range []string{"mihomo", "system", "listeners"} {
				if _, exists := resp[key]; !exists {
					t.Fatalf("expected aggregator to contain %q", key)
				}
			}
		}
	}

	_ = os.RemoveAll("/tmp/3m-ui-router-test")
}

func TestCORSConfiguredOrigin(t *testing.T) {
	cfg := &config.Config{}
	cfg.Server.Mode = "debug"
	cfg.Security.CORSOrigins = []string{"https://panel.example.com"}
	cfg.JWT.Secret = "test-secret"

	r := router.SetupRouter(cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/health", nil)
	req.Header.Set("Origin", "https://panel.example.com")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://panel.example.com" {
		t.Fatalf("expected reflected origin, got %q", got)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/health", nil)
	req2.Header.Set("Origin", "https://evil.example.com")
	r.ServeHTTP(w2, req2)
	if got := w2.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no allow-origin for disallowed origin, got %q", got)
	}
}
