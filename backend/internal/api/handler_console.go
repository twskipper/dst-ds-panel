package api

import (
	"encoding/json"
	"net/http"

	"dst-ds-panel/internal/model"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) SendConsoleCommand(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "clusterID")
	shardName := chi.URLParam(r, "shard")

	cluster, err := h.store.GetCluster(clusterID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	shard := findShard(cluster, shardName)
	if shard == nil || shard.ContainerID == "" || shard.Status != model.StatusRunning {
		writeError(w, http.StatusBadRequest, "shard is not running")
		return
	}

	var body struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Command == "" {
		writeError(w, http.StatusBadRequest, "command is required")
		return
	}

	if err := h.docker.ExecCommand(r.Context(), shard.ContainerID, body.Command); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "sent",
		"command": body.Command,
	})
}
