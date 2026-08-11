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

	user.InitService(
		database.GlobalDB,
	)

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

	if value := os.Getenv("THREE_M_UI_CONFIG"); value != "" {
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


	fileServer := http.FileServer(
		http.FS(staticFS),
	)


	r.NoRoute(func(c *gin.Context) {


		path := c.Request.URL.Path


		if path == "/" {

			c.Data(
				http.StatusOK,
				"text/html; charset=utf-8",
				mustReadFile(staticFS,"index.html"),
			)

			return
		}


		f, err := staticFS.Open(path[1:])


		if err == nil {

			defer f.Close()

			fileServer.ServeHTTP(c.Writer,c.Request)

			return
		}


		c.Data(
			http.StatusOK,
			"text/html; charset=utf-8",
			mustReadFile(staticFS,"index.html"),
		)

	})
}



func mustReadFile(
	fsys fs.FS,
	name string,
) []byte {


	data,err := fs.ReadFile(fsys,name)


	if err != nil {

		log.Printf(
			"read frontend %s failed:%v",
			name,
			err,
		)

		return []byte("3m-ui frontend unavailable")
	}


	return data
}
