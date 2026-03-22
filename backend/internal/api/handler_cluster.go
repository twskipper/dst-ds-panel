package api

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dst-ds-panel/internal/dst"
	"dst-ds-panel/internal/model"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListClusters(w http.ResponseWriter, r *http.Request) {
	clusters := h.store.ListClusters()
	writeJSON(w, http.StatusOK, clusters)
}

func (h *Handler) GetCluster(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cluster)
}

type CreateClusterRequest struct {
	Name        string              `json:"name"`
	Config      model.ClusterConfig `json:"config"`
	EnableCaves *bool               `json:"enableCaves"`
}

func (h *Handler) CreateCluster(w http.ResponseWriter, r *http.Request) {
	var req CreateClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	id := sanitizeID(req.Name)
	dirName := id
	clusterDir := filepath.Join(h.dataDir, "clusters", dirName)

	if _, err := os.Stat(clusterDir); err == nil {
		writeError(w, http.StatusConflict, "cluster directory already exists")
		return
	}

	cavesEnabled := req.EnableCaves == nil || *req.EnableCaves // default true

	if err := dst.InitClusterDir(clusterDir, &req.Config, cavesEnabled); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("init cluster: %v", err))
		return
	}

	shards := []model.Shard{
		{Name: "Master", Status: model.StatusStopped, Config: model.ShardConfig{IsMaster: true, ServerPort: 10999, MasterPort: 27018}},
	}
	if cavesEnabled {
		shards = append(shards, model.Shard{Name: "Caves", Status: model.StatusStopped, Config: model.ShardConfig{IsMaster: false, ServerPort: 10998, MasterPort: 27019}})
	}

	cluster := model.Cluster{
		ID:        id,
		Name:      req.Name,
		DirName:   dirName,
		Status:    model.StatusStopped,
		Config:    req.Config,
		Shards:    shards,
		CreatedAt: time.Now(),
	}

	if err := h.store.SaveCluster(cluster); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, cluster)
}

func (h *Handler) UpdateClusterConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var config model.ClusterConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	clusterDir := filepath.Join(h.dataDir, "clusters", cluster.DirName)
	if err := dst.WriteClusterConfig(clusterDir, &config); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cluster.Config = config
	if err := h.store.SaveCluster(*cluster); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cluster)
}

func (h *Handler) CloneCluster(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	newID := sanitizeID(body.Name)
	srcDir := filepath.Join(h.dataDir, "clusters", cluster.DirName)
	destDir := filepath.Join(h.dataDir, "clusters", newID)

	if _, err := os.Stat(destDir); err == nil {
		writeError(w, http.StatusConflict, "cluster with this name already exists")
		return
	}

	if err := copyDir(srcDir, destDir); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Remove cluster_token.txt from clone (user should set their own)
	os.Remove(filepath.Join(destDir, "cluster_token.txt"))

	newCluster := model.Cluster{
		ID:      newID,
		Name:    body.Name,
		DirName: newID,
		Status:  model.StatusStopped,
		Config:  cluster.Config,
		Shards:  make([]model.Shard, len(cluster.Shards)),
		CreatedAt: time.Now(),
	}
	for i, s := range cluster.Shards {
		newCluster.Shards[i] = model.Shard{
			Name:   s.Name,
			Status: model.StatusStopped,
			Config: s.Config,
		}
	}
	newCluster.Config.ClusterName = body.Name

	if err := h.store.SaveCluster(newCluster); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, newCluster)
}

func (h *Handler) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// Stop containers if running
	if cluster.Status == model.StatusRunning {
		for _, shard := range cluster.Shards {
			if shard.ContainerID != "" {
				_ = h.shardMgr.StopShard(r.Context(), shard.ContainerID)
			}
		}
	}

	clusterDir := filepath.Join(h.dataDir, "clusters", cluster.DirName)
	if err := os.RemoveAll(clusterDir); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.store.DeleteCluster(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ImportCluster(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(100 << 20) // 100MB max

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file")
		return
	}
	defer file.Close()

	name := r.FormValue("name")
	if name == "" {
		name = strings.TrimSuffix(header.Filename, ".zip")
	}

	id := sanitizeID(name)
	clusterDir := filepath.Join(h.dataDir, "clusters", id)
	if err := os.MkdirAll(clusterDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Save uploaded file to temp
	tmpFile, err := os.CreateTemp("", "dst-import-*.zip")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, file); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tmpFile.Close()

	// Unzip to temp dir first, then find the actual cluster root
	tmpExtract, err := os.MkdirTemp("", "dst-extract-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer os.RemoveAll(tmpExtract)

	if err := unzip(tmpFile.Name(), tmpExtract); err != nil {
		os.RemoveAll(clusterDir)
		writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to extract zip: %v", err))
		return
	}

	// Find actual cluster root: look for Master/ or cluster.ini
	clusterRoot := findClusterRoot(tmpExtract)
	if clusterRoot == "" {
		os.RemoveAll(clusterDir)
		writeError(w, http.StatusBadRequest, "could not find cluster data in zip (no Master/ directory or cluster.ini found)")
		return
	}

	// Copy from detected root to the cluster dir
	if err := copyDir(clusterRoot, clusterDir); err != nil {
		os.RemoveAll(clusterDir)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to copy cluster: %v", err))
		return
	}

	fmt.Printf("Import: extracted cluster from %s to %s\n", clusterRoot, clusterDir)

	// Read config from imported cluster.ini
	config, err := dst.ReadClusterConfig(clusterDir)
	if err != nil {
		config = &model.ClusterConfig{
			ClusterName: name,
			GameMode:    "survival",
			MaxPlayers:  6,
		}
	}

	// Auto-detect mods from Master/modoverrides.lua and generate mods_setup.lua
	masterDir := filepath.Join(clusterDir, "Master")
	mods, _ := dst.ReadModOverrides(masterDir)
	if len(mods) > 0 {
		_ = dst.WriteModsSetup(clusterDir, mods)
		fmt.Printf("Import: detected %d mods from modoverrides.lua\n", len(mods))
	}

	// Detect shards
	shards := []model.Shard{
		{Name: "Master", Status: model.StatusStopped, Config: model.ShardConfig{IsMaster: true, ServerPort: 10999}},
	}
	if _, err := os.Stat(filepath.Join(clusterDir, "Caves")); err == nil {
		shards = append(shards, model.Shard{Name: "Caves", Status: model.StatusStopped, Config: model.ShardConfig{IsMaster: false, ServerPort: 10998}})
	}

	cluster := model.Cluster{
		ID:        id,
		Name:      name,
		DirName:   id,
		Status:    model.StatusStopped,
		Config:    *config,
		Shards:    shards,
		CreatedAt: time.Now(),
	}

	if err := h.store.SaveCluster(cluster); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, cluster)
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// findClusterRoot searches for the actual cluster directory inside an extracted zip.
// It looks for a directory containing Master/ or cluster.ini, handling cases where
// the zip has a root directory prefix (e.g., Cluster_2/Master/ instead of Master/).
func findClusterRoot(dir string) string {
	// Check if this dir itself is the cluster root
	if isClusterDir(dir) {
		return dir
	}

	// Check immediate subdirectories (one level of nesting)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subDir := filepath.Join(dir, e.Name())
		if isClusterDir(subDir) {
			return subDir
		}
	}

	return ""
}

func isClusterDir(dir string) bool {
	// A cluster dir has Master/ or cluster.ini
	if _, err := os.Stat(filepath.Join(dir, "Master")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "cluster.ini")); err == nil {
		return true
	}
	return false
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

func sanitizeID(name string) string {
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	var result []byte
	for _, c := range []byte(id) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result = append(result, c)
		}
	}
	if len(result) == 0 {
		return fmt.Sprintf("cluster-%d", time.Now().Unix())
	}
	return string(result)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
