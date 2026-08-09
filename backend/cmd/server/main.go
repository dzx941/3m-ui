package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/dzx941/3m-ui/backend/internal/config"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/router"
)

func main() {
	configPath := flag.String("config", "backend/config/config.yaml", "path to config file")
	flag.Parse()

	log.Printf("Starting 3m-ui backend server...")

	// 1. Load config
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	log.Printf("Configuration loaded successfully from %s", *configPath)

	// 2. Initialize database
	db, err := database.InitDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	_ = db
	log.Printf("Database initialized and migrated successfully at %s", cfg.Database.Path)

	// 3. Setup router
	r := router.SetupRouter(cfg)

	// 4. Start HTTP Server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("HTTP Server is listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
