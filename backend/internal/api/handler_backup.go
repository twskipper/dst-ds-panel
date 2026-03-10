package api

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) BackupCluster(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(clusterID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	clusterDir := filepath.Join(h.dataDir, "clusters", cluster.DirName)
	if _, err := os.Stat(clusterDir); os.IsNotExist(err) {
		writeError(w, http.StatusNotFound, "cluster directory not found")
		return
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.zip", cluster.DirName, timestamp)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	zw := zip.NewWriter(w)
	defer zw.Close()

	err = filepath.Walk(clusterDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from cluster dir
		relPath, err := filepath.Rel(clusterDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Use forward slashes in zip
		zipPath := strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		if info.IsDir() {
			_, err := zw.Create(zipPath + "/")
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipPath
		header.Method = zip.Deflate

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		// At this point headers are already sent, just log
		fmt.Printf("backup error: %v\n", err)
	}
}
