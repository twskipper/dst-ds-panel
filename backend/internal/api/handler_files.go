package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Allowed file paths relative to cluster directory
var allowedFiles = map[string]bool{
	"cluster.ini":                  true,
	"adminlist.txt":                true,
	"blocklist.txt":                true,
	"whitelist.txt":                true,
	"Master/server.ini":            true,
	"Master/modoverrides.lua":      true,
	"Master/worldgenoverride.lua":  true,
	"Master/leveldataoverride.lua": true,
	"Caves/server.ini":             true,
	"Caves/modoverrides.lua":       true,
	"Caves/worldgenoverride.lua":   true,
	"Caves/leveldataoverride.lua":  true,
}

// Allowed directories for new files
var allowedDirs = map[string]bool{
	".":      true,
	"Master": true,
	"Caves":  true,
}

// Allowed extensions for new files
var allowedExts = map[string]bool{
	".ini": true,
	".lua": true,
	".txt": true,
}

func isAllowedFile(path string) bool {
	normalized := filepath.ToSlash(filepath.Clean(path))

	// Check whitelist first
	if allowedFiles[normalized] {
		return true
	}

	// Allow new files in permitted directories with permitted extensions
	if strings.Contains(normalized, "..") {
		return false
	}
	dir := filepath.ToSlash(filepath.Dir(normalized))
	ext := filepath.Ext(normalized)
	return allowedDirs[dir] && allowedExts[ext]
}

func langFromPath(path string) string {
	if strings.HasSuffix(path, ".lua") {
		return "lua"
	}
	if strings.HasSuffix(path, ".ini") {
		return "ini"
	}
	return "plaintext"
}

func (h *Handler) ReadFile(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "clusterID")
	filePath := r.URL.Query().Get("path")

	cluster, err := h.store.GetCluster(clusterID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if !isAllowedFile(filePath) {
		writeError(w, http.StatusBadRequest, "file not allowed: "+filePath)
		return
	}

	fullPath := filepath.Join(h.dataDir, "clusters", cluster.DirName, filepath.FromSlash(filePath))
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, map[string]string{"content": "", "path": filePath})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"content": string(data), "path": filePath})
}

func (h *Handler) WriteFile(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "clusterID")
	filePath := r.URL.Query().Get("path")

	cluster, err := h.store.GetCluster(clusterID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if !isAllowedFile(filePath) {
		writeError(w, http.StatusBadRequest, "file not allowed: "+filePath)
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fullPath := filepath.Join(h.dataDir, "clusters", cluster.DirName, filepath.FromSlash(filePath))
	os.MkdirAll(filepath.Dir(fullPath), 0755)

	if err := os.WriteFile(fullPath, []byte(body.Content), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "saved", "path": filePath})
}

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(clusterID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	clusterDir := filepath.Join(h.dataDir, "clusters", cluster.DirName)
	var files []map[string]string

	// List whitelisted files
	for path := range allowedFiles {
		fullPath := filepath.Join(clusterDir, filepath.FromSlash(path))
		exists := false
		if _, err := os.Stat(fullPath); err == nil {
			exists = true
		}
		files = append(files, map[string]string{
			"path":   path,
			"exists": boolStr(exists),
			"lang":   langFromPath(path),
		})
	}

	// Also scan for extra .lua/.ini/.txt files not in whitelist
	for dir := range allowedDirs {
		scanDir := filepath.Join(clusterDir, dir)
		entries, err := os.ReadDir(scanDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := filepath.Ext(e.Name())
			if !allowedExts[ext] {
				continue
			}
			relPath := e.Name()
			if dir != "." {
				relPath = dir + "/" + e.Name()
			}
			// Skip if already in whitelist
			if allowedFiles[relPath] {
				continue
			}
			files = append(files, map[string]string{
				"path":   relPath,
				"exists": "true",
				"lang":   langFromPath(relPath),
			})
		}
	}

	writeJSON(w, http.StatusOK, files)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
