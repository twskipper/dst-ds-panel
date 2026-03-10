package service

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func StartAutoBackup(dataDir string, intervalHours int) {
	if intervalHours <= 0 {
		return
	}
	log.Printf("Auto-backup enabled: every %d hour(s)", intervalHours)

	go func() {
		ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			backupAll(dataDir)
		}
	}()
}

func backupAll(dataDir string) {
	clustersDir := filepath.Join(dataDir, "clusters")
	backupDir := filepath.Join(dataDir, "backups")
	os.MkdirAll(backupDir, 0755)

	entries, err := os.ReadDir(clustersDir)
	if err != nil {
		log.Printf("Auto-backup: could not read clusters dir: %v", err)
		return
	}

	timestamp := time.Now().Format("20060102-150405")

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		clusterDir := filepath.Join(clustersDir, e.Name())
		zipName := fmt.Sprintf("%s-%s.zip", e.Name(), timestamp)
		zipPath := filepath.Join(backupDir, zipName)

		if err := createZip(clusterDir, zipPath); err != nil {
			log.Printf("Auto-backup failed for %s: %v", e.Name(), err)
			continue
		}
		log.Printf("Auto-backup: %s -> %s", e.Name(), zipName)
	}

	// Clean old backups: keep last 10 per cluster
	cleanOldBackups(backupDir, entries)
}

func createZip(srcDir, destZip string) error {
	f, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(srcDir, path)
		if relPath == "." {
			return nil
		}
		zipPath := strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		if info.IsDir() {
			_, err := zw.Create(zipPath + "/")
			return err
		}

		header, _ := zip.FileInfoHeader(info)
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
}

func cleanOldBackups(backupDir string, clusters []os.DirEntry) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}
	for _, cluster := range clusters {
		prefix := cluster.Name() + "-"
		var matching []string
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), prefix) && strings.HasSuffix(e.Name(), ".zip") {
				matching = append(matching, e.Name())
			}
		}
		if len(matching) > 10 {
			for _, name := range matching[:len(matching)-10] {
				os.Remove(filepath.Join(backupDir, name))
			}
		}
	}
}
