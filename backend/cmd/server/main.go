package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/listener"
	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	"github.com/dzx941/3m-ui/backend/internal/node"
	"github.com/dzx941/3m-ui/backend/internal/router"
	"github.com/dzx941/3m-ui/backend/internal/security"
	"github.com/dzx941/3m-ui/backend/internal/traffic"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"github.com/gin-gonic/gin"
)

//go:embed web/dist/*
var frontendFiles embed.FS

func main() {
	if len(os.Args) > 1 &&
		(os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Println(versionString())
		return
	}

	configPath := defaultConfigPath()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if _, err := database.InitDB(cfg.Database.Path); err != nil {
		log.Fatalf("initialize database: %v", err)
	}

	security.InitCredentialKey(cfg.Security.CredentialKey)

	mihomo.InitService(cfg)

	listener.InitService(
		database.GlobalDB,
		cfg.Mihomo.Config,
	)

	node.InitService(
		database.GlobalDB,
		cfg.Mihomo.Config,
	)

	user.InitService(database.GlobalDB)

	traffic.InitGlobalService()

	r := router.SetupRouter(cfg)

	mountFrontend(r)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	log.Printf("3m-ui listening on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("run server: %v", err)
	}
}


func defaultConfigPath() string {
	if value := os.Getenv("3M_UI_CONFIG"); value != "" {
		return value
	}

	if _, err := os.Stat("/etc/3m-ui/config.yaml"); err == nil {
		return "/etc/3m-ui/config.yaml"
	}

	return "backend/config/config.yaml"
}


func mountFrontend(r *gin.Engine) {
	staticFS, err := fs.Sub(frontendFiles, "web/dist")
	if err != nil {
		log.Printf("frontend assets unavailable: %v", err)
		return
	}

	// Disable automatic redirect problems
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false


	// Explicit root route
	// Prevent Gin 301 redirect loop on /
	r.GET("/", func(c *gin.Context) {
		c.FileFromFS(
			"index.html",
			http.FS(staticFS),
		)
	})


	// SPA fallback
	r.NoRoute(func(c *gin.Context) {

		path := c.Request.URL.Path

		// API should never fallback to frontend
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": "api endpoint not found",
				},
			)
			return
		}


		// Frontend routes:
		// /users
		// /dashboard
		// /listeners
		// etc.
		if filepath.Ext(path) == "" {

			c.FileFromFS(
				"index.html",
				http.FS(staticFS),
			)

			return
		}


		// Static assets:
		// /assets/*.js
		// /assets/*.css
		// favicon.ico
		err := c.FileFromFS(
			path[1:],
			http.FS(staticFS),
		)

		if err != nil {
			c.Status(http.StatusNotFound)
		}
	})
}
