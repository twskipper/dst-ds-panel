package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"dst-ds-panel/internal/model"
	"dst-ds-panel/internal/service"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) StartCluster(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if cluster.Status == model.StatusRunning {
		writeError(w, http.StatusConflict, "cluster is already running")
		return
	}

	// Check for cluster token
	tokenPath := filepath.Join(h.dataDir, "clusters", cluster.DirName, "cluster_token.txt")
	tokenData, err := os.ReadFile(tokenPath)
	if err != nil || strings.TrimSpace(string(tokenData)) == "" {
		writeError(w, http.StatusBadRequest, "Cluster token is required. Set it in Overview → Cluster Token before starting.")
		return
	}


	// Auto-pull runtime image if not present
	if err := h.shardMgr.EnsureImage(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to pull Docker image: %v", err))
		return
	}
	cluster.Status = model.StatusStarting
	h.store.SaveCluster(*cluster)

	for i, shard := range cluster.Shards {
		containerID, err := h.shardMgr.StartShard(r.Context(), cluster.DirName, shard.Name)
		if err != nil {
			cluster.Status = model.StatusError
			h.store.SaveCluster(*cluster)
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("start %s: %v", shard.Name, err))
			return
		}
		cluster.Shards[i].ContainerID = containerID
		cluster.Shards[i].Status = model.StatusRunning
	}

	cluster.Status = model.StatusRunning
	h.store.SaveCluster(*cluster)
	service.NotifyServerStarted(cluster.Name)
	writeJSON(w, http.StatusOK, cluster)
}

func (h *Handler) StopCluster(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	for i, shard := range cluster.Shards {
		if shard.ContainerID != "" {
			if err := h.shardMgr.StopShard(r.Context(), shard.ContainerID); err != nil {
				// Log but continue stopping other shards
				fmt.Printf("warning: failed to stop %s: %v\n", shard.Name, err)
			}
			cluster.Shards[i].ContainerID = ""
			cluster.Shards[i].Status = model.StatusStopped
		}
	}

	cluster.Status = model.StatusStopped
	h.store.SaveCluster(*cluster)
	service.NotifyServerStopped(cluster.Name)
	writeJSON(w, http.StatusOK, cluster)
}

func (h *Handler) RestartCluster(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// Stop all shards
	for _, shard := range cluster.Shards {
		if shard.ContainerID != "" {
			_ = h.shardMgr.StopShard(r.Context(), shard.ContainerID)
		}
	}

	// Start all shards
	for i, shard := range cluster.Shards {
		containerID, err := h.shardMgr.StartShard(r.Context(), cluster.DirName, shard.Name)
		if err != nil {
			cluster.Status = model.StatusError
			h.store.SaveCluster(*cluster)
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("restart %s: %v", shard.Name, err))
			return
		}
		cluster.Shards[i].ContainerID = containerID
		cluster.Shards[i].Status = model.StatusRunning
	}

	cluster.Status = model.StatusRunning
	h.store.SaveCluster(*cluster)
	writeJSON(w, http.StatusOK, cluster)
}
