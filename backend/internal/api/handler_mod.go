package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"dst-ds-panel/internal/dst"
	"dst-ds-panel/internal/model"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListMods(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(clusterID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	masterDir := filepath.Join(h.dataDir, "clusters", cluster.DirName, "Master")
	mods, err := dst.ReadModOverrides(masterDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if mods == nil {
		mods = []model.Mod{}
	}

	writeJSON(w, http.StatusOK, mods)
}

func (h *Handler) UpdateMods(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(clusterID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var mods []model.Mod
	if err := json.NewDecoder(r.Body).Decode(&mods); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	clusterDir := filepath.Join(h.dataDir, "clusters", cluster.DirName)
	masterDir := filepath.Join(clusterDir, "Master")
	cavesDir := filepath.Join(clusterDir, "Caves")

	// Write modoverrides.lua to both shards
	if err := dst.WriteModOverrides(masterDir, mods); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := dst.WriteModOverrides(cavesDir, mods); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Write mods_setup.lua
	if err := dst.WriteModsSetup(clusterDir, mods); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, mods)
}
