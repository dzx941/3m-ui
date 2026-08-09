package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/router"
)

func TestHealthCheck(t *testing.T) {
	cfg := &config.Config{}
	cfg.Server.Mode = "debug"

	r := router.SetupRouter(cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if response["status"] != "ok" {
		t.Fatalf("expected status 'ok', got '%s'", response["status"])
	}
}
