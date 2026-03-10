package api

import (
	"net/http"

	"dst-ds-panel/internal/model"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "clusterID")
	shardName := chi.URLParam(r, "shard")

	cluster, err := h.store.GetCluster(clusterID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	shard := findShard(cluster, shardName)
	if shard == nil || shard.ContainerID == "" {
		writeError(w, http.StatusNotFound, "shard not running")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	lines, err := h.docker.StreamLogsLines(r.Context(), shard.ContainerID)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}

	for line := range lines {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			return
		}
	}
}

func (h *Handler) StreamStats(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "clusterID")
	shardName := chi.URLParam(r, "shard")

	cluster, err := h.store.GetCluster(clusterID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	shard := findShard(cluster, shardName)
	if shard == nil || shard.ContainerID == "" {
		writeError(w, http.StatusNotFound, "shard not running")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	statsCh, err := h.docker.StreamStats(r.Context(), shard.ContainerID)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}

	for stat := range statsCh {
		if err := conn.WriteJSON(stat); err != nil {
			return
		}
	}
}

func findShard(cluster *model.Cluster, name string) *model.Shard {
	for i := range cluster.Shards {
		if cluster.Shards[i].Name == name {
			return &cluster.Shards[i]
		}
	}
	return nil
}
