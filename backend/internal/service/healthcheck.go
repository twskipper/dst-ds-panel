package service

import (
	"context"
	"log"
	"time"

	"dst-ds-panel/internal/manager"
	"dst-ds-panel/internal/model"
	"dst-ds-panel/internal/store"
)

func StartHealthCheck(mgr manager.ShardManager, s *store.Store, intervalSeconds int) {
	if intervalSeconds <= 0 {
		intervalSeconds = 30
	}
	log.Printf("Health check enabled: every %ds", intervalSeconds)

	go func() {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			checkShardHealth(mgr, s)
		}
	}()
}

func checkShardHealth(mgr manager.ShardManager, s *store.Store) {
	ctx := context.Background()
	running, err := mgr.ListRunningShards(ctx)
	if err != nil {
		return
	}

	clusters := s.ListClusters()
	for _, cluster := range clusters {
		changed := false
		allStopped := true

		for i, shard := range cluster.Shards {
			key := cluster.DirName + "/" + shard.Name
			_, isRunning := running[key]

			if shard.Status == model.StatusRunning && !isRunning {
				// Container died
				cluster.Shards[i].Status = model.StatusStopped
				cluster.Shards[i].ContainerID = ""
				changed = true
				log.Printf("Health check: %s/%s container stopped unexpectedly", cluster.Name, shard.Name)
				NotifyServerError(cluster.Name, shard.Name+" shard stopped unexpectedly")
			} else if shard.Status == model.StatusStopped && isRunning {
				// Container running but store says stopped (edge case)
				cluster.Shards[i].Status = model.StatusRunning
				cluster.Shards[i].ContainerID = running[key]
				changed = true
			}

			if cluster.Shards[i].Status == model.StatusRunning {
				allStopped = false
			}
		}

		if changed {
			if allStopped && cluster.Status == model.StatusRunning {
				cluster.Status = model.StatusStopped
			} else if !allStopped && cluster.Status == model.StatusStopped {
				cluster.Status = model.StatusRunning
			}
			s.SaveCluster(cluster)
		}
	}
}
