package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"dst-ds-panel/internal/api"
	"dst-ds-panel/internal/config"
	"dst-ds-panel/internal/docker"
	"dst-ds-panel/internal/model"
	"dst-ds-panel/internal/service"
	"dst-ds-panel/internal/store"
)

//go:embed all:frontend
var frontendFS embed.FS

//go:embed world-settings.json
var defaultWorldSettings []byte

func reconcileStatus(dockerMgr *docker.Manager, s *store.Store) {
	ctx := context.Background()
	running, err := dockerMgr.ListRunningShards(ctx)
	if err != nil {
		log.Printf("Warning: could not list Docker containers: %v", err)
		return
	}

	clusters := s.ListClusters()
	for _, cluster := range clusters {
		anyRunning := false
		for i, shard := range cluster.Shards {
			key := cluster.DirName + "/" + shard.Name
			if containerID, ok := running[key]; ok {
				cluster.Shards[i].ContainerID = containerID
				cluster.Shards[i].Status = model.StatusRunning
				anyRunning = true
			} else {
				cluster.Shards[i].ContainerID = ""
				cluster.Shards[i].Status = model.StatusStopped
			}
		}
		if anyRunning {
			cluster.Status = model.StatusRunning
		} else {
			cluster.Status = model.StatusStopped
		}
		s.SaveCluster(cluster)
	}
	log.Printf("Reconciled %d clusters with Docker state (%d running containers)", len(clusters), len(running))
}

func main() {
	// CLI flags
	dumpWorldSettings := flag.Bool("dump-world-settings", false, "Dump embedded world-settings.json to stdout and exit")
	worldSettingsFile := flag.String("world-settings", "", "Path to custom world-settings.json file")
	flag.Parse()

	if *dumpWorldSettings {
		fmt.Print(string(defaultWorldSettings))
		os.Exit(0)
	}

	// Load world settings: custom file or embedded default
	worldSettingsData := defaultWorldSettings
	if *worldSettingsFile != "" {
		data, err := os.ReadFile(*worldSettingsFile)
		if err != nil {
			log.Fatalf("Failed to read world settings file: %v", err)
		}
		worldSettingsData = data
		log.Printf("Using custom world settings: %s", *worldSettingsFile)
	}

	// Load config: try current dir, then parent dir (when run from backend/)
	configPath := "config.json"
	if v := os.Getenv("CONFIG_FILE"); v != "" {
		configPath = v
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		cfg, _ = config.Load(filepath.Join("..", configPath))
		if cfg == nil {
			cfg = config.DefaultConfig()
		}
	}

	// Resolve data directory
	dataDir, err := filepath.Abs(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}
	// When running from backend/ subdir during development, check if ../data exists
	// Only switch if current resolved dataDir has no data yet AND parent does
	if _, e := os.Stat(dataDir); os.IsNotExist(e) {
		if alt, err2 := filepath.Abs(filepath.Join("..", cfg.DataDir)); err2 == nil {
			if _, e2 := os.Stat(alt); e2 == nil {
				dataDir = alt
			}
		}
	}

	os.MkdirAll(filepath.Join(dataDir, "clusters"), 0755)

	dockerMgr, err := docker.NewManager(dataDir, cfg.ImageName, cfg.Platform)
	if err != nil {
		log.Fatal("Docker not available:", err)
	}
	defer dockerMgr.Close()

	s, err := store.New(filepath.Join(dataDir, "store.json"))
	if err != nil {
		log.Fatal("Store init failed:", err)
	}

	// Reconcile cluster status with actual Docker containers
	reconcileStatus(dockerMgr, s)

	// Start services
	service.InitDiscord(cfg.DiscordWebhook)
	service.StartAutoBackup(dataDir, cfg.BackupInterval)
	service.StartHealthCheck(dockerMgr, s, 30)

	// Embedded frontend
	frontendContent, fsErr := fs.Sub(frontendFS, "frontend")
	if fsErr != nil {
		log.Fatal("Failed to load embedded frontend:", fsErr)
	}

	h := api.NewHandler(dockerMgr, s, dataDir)
	router := api.NewRouter(h, cfg.Auth, frontendContent, worldSettingsData)

	log.Printf("Config: image=%s platform=%s auth_user=%s", cfg.ImageName, cfg.Platform, cfg.Auth.Username)
	log.Printf("Data directory: %s", dataDir)
	log.Printf("Starting DST DS Panel on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
