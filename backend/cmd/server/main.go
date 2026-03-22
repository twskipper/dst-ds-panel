package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"runtime"

	"dst-ds-panel/internal/api"
	"dst-ds-panel/internal/config"
	"dst-ds-panel/internal/manager"
	"dst-ds-panel/internal/model"
	"dst-ds-panel/internal/service"
	"dst-ds-panel/internal/store"
)

//go:embed all:frontend
var frontendFS embed.FS

func reconcileStatus(mgr manager.ShardManager, s *store.Store) {
	ctx := context.Background()
	running, err := mgr.ListRunningShards(ctx)
	if err != nil {
		log.Printf("Warning: could not list running shards: %v", err)
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
	log.Printf("Reconciled %d clusters (%d running shards)", len(clusters), len(running))
}

func main() {
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
	if _, e := os.Stat(dataDir); os.IsNotExist(e) {
		if alt, err2 := filepath.Abs(filepath.Join("..", cfg.DataDir)); err2 == nil {
			if _, e2 := os.Stat(alt); e2 == nil {
				dataDir = alt
			}
		}
	}

	os.MkdirAll(filepath.Join(dataDir, "clusters"), 0755)

	// Determine runtime mode: native (no Docker) or docker
	mode := os.Getenv("DST_MODE")
	if mode == "" {
		if runtime.GOOS == "windows" {
			mode = "native"
		} else {
			mode = "docker"
		}
	}

	var shardMgr manager.ShardManager
	if mode == "native" {
		shardMgr = manager.NewProcessManager(dataDir)
		log.Println("Running in native mode (no Docker)")
	} else {
		dockerMgr, err := manager.NewDockerManager(dataDir, cfg.ImageName, cfg.Platform)
		if err != nil {
			log.Fatal("Docker not available:", err)
		}
		shardMgr = dockerMgr
		log.Println("Running in Docker mode")
	}
	defer shardMgr.Close()

	s, err := store.New(filepath.Join(dataDir, "store.json"))
	if err != nil {
		log.Fatal("Store init failed:", err)
	}

	// Reconcile cluster status with running shards
	reconcileStatus(shardMgr, s)

	// Start services
	service.InitDiscord(cfg.DiscordWebhook)
	service.StartAutoBackup(dataDir, cfg.BackupInterval)
	service.StartHealthCheck(shardMgr, s, 30)

	// Embedded frontend
	frontendContent, fsErr := fs.Sub(frontendFS, "frontend")
	if fsErr != nil {
		log.Fatal("Failed to load embedded frontend:", fsErr)
	}

	h := api.NewHandler(shardMgr, s, dataDir, mode)
	router := api.NewRouter(h, cfg.Auth, frontendContent)

	log.Printf("Config: mode=%s image=%s auth_user=%s", mode, cfg.ImageName, cfg.Auth.Username)
	log.Printf("Data directory: %s", dataDir)
	log.Printf("Starting DST DS Panel on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
