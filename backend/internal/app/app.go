package app

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dzx941/3m-ui/backend/internal/auth"
	"github.com/dzx941/3m-ui/backend/internal/config"
	dbconfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/router"
	"github.com/dzx941/3m-ui/backend/internal/security"
	"github.com/gin-gonic/gin"
)

func Run(frontendFS fs.FS) error {
	configPath := defaultConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil { return fmt.Errorf("load config: %w", err) }
	db, err := database.InitDB(cfg.Database.Path)
	if err != nil { return fmt.Errorf("initialize database: %w", err) }
	if created, username, password, err := auth.EnsureAdmin(db, cfg.Database.Path); err != nil {
		return fmt.Errorf("initialize administrator: %w", err)
	} else if created {
		log.Printf("initial administrator created: username=%s", username)
		passwordFile := filepath.Join(filepath.Dir(cfg.Database.Path), ".initial_admin_password")
		if err := os.WriteFile(passwordFile, []byte(password+"\n"), 0600); err != nil { log.Printf("warning: could not write initial admin password file: %v", err) } else { log.Printf("initial administrator password saved to %s", passwordFile) }
	}

	security.InitCredentialKey(cfg.Security.CredentialKey)
	container := NewContainer(db, cfg)

	dbconfig.CredentialProvider = func() (map[uint][]dbconfig.Credential, error) {
		if container.User == nil { return map[uint][]dbconfig.Credential{}, nil }
		provided, err := container.User.ActiveCredentialsByListener(); if err != nil { return nil, err }
		result := make(map[uint][]dbconfig.Credential, len(provided))
		for listenerID, credentials := range provided {
			converted := make([]dbconfig.Credential, 0, len(credentials))
			for _, credential := range credentials { converted = append(converted, dbconfig.Credential{Username: credential.Username, Password: credential.Password, UUID: credential.UUID}) }
			result[listenerID] = converted
		}
		return result, nil
	}

	generatedConfig, err := container.ConfigEngine.GenerateFinalConfig(); if err != nil { return fmt.Errorf("generate Mihomo configuration: %w", err) }
	if container.Mihomo == nil { return fmt.Errorf("initialize Mihomo service: service is nil") }
	if err := container.Mihomo.SaveConfig(generatedConfig); err != nil { return fmt.Errorf("validate Mihomo configuration: %w", err) }
	if err := container.Mihomo.StartMihomo(); err != nil { return fmt.Errorf("start Mihomo core: %w", err) }
	log.Printf("Mihomo core started successfully")

	if container.Scheduler != nil {
		container.Scheduler.Start()
		defer container.Scheduler.Stop()
	}

	routerDeps := router.Dependencies{
		DB: container.DB, Config: container.Config, Mihomo: container.Mihomo, Listener: container.Listener,
		Node: container.Node, User: container.User, System: container.System, Traffic: container.Traffic,
		Collector: container.Collector, ConfigEngine: container.ConfigEngine,
	}
	router.ConfigureDependencies(routerDeps)
	r := router.SetupRouter(cfg, routerDeps)
	mountFrontend(r, frontendFS)
	addr := fmt.Sprintf(":%d", cfg.Server.Port); log.Printf("3m-ui listening on %s", addr)
	if err := r.Run(addr); err != nil { return fmt.Errorf("run server: %w", err) }
	return nil
}

func defaultConfigPath() string {
	if value := os.Getenv("THREE_M_UI_CONFIG"); value != "" { return value }
	if _, err := os.Stat("/etc/3m-ui/config.yaml"); err == nil { return "/etc/3m-ui/config.yaml" }
	return "backend/config/config.yaml"
}

func mountFrontend(r *gin.Engine, frontendFS fs.FS) {
	staticFS, err := fs.Sub(frontendFS, "web/dist"); if err != nil { log.Printf("frontend assets unavailable: %v", err); return }
	fileServer := http.FileServer(http.FS(staticFS)); r.RedirectTrailingSlash = false; r.RedirectFixedPath = false
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) >= 4 && path[:4] == "/api" { c.Status(http.StatusNotFound); return }
		if path == "/" { c.Data(http.StatusOK, "text/html; charset=utf-8", mustReadFile(staticFS, "index.html")); return }
		f, err := staticFS.Open(path[1:]); if err == nil { defer f.Close(); fileServer.ServeHTTP(c.Writer, c.Request); return }
		c.Data(http.StatusOK, "text/html; charset=utf-8", mustReadFile(staticFS, "index.html"))
	})
}

func mustReadFile(fsys fs.FS, name string) []byte {
	data, err := fs.ReadFile(fsys, name); if err != nil { log.Printf("read frontend %s failed: %v", name, err); return []byte("3m-ui frontend unavailable") }; return data
}
