package main

import (
	"log"
	"os"

	_ "github.com/yurisachan16-creator/Memory-system/backend/docs"
	appconfig "github.com/yurisachan16-creator/Memory-system/backend/internal/config"
	"github.com/yurisachan16-creator/Memory-system/backend/internal/server"
)

// @title Memory System API
// @version 1.0
// @description Backend API for a simplified AI memory system.
// @BasePath /api/v1
func main() {
	configPath := os.Getenv("CONFIG_PATH")
	cfg, err := appconfig.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	app, err := server.New(cfg)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
