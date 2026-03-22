package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (h *Handler) ImageStatus(w http.ResponseWriter, r *http.Request) {
	exists, err := h.docker.ImageExists(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	dstVersion := ""
	versionFile := filepath.Join(h.dataDir, "dst_server", "version.txt")
	if data, err := os.ReadFile(versionFile); err == nil {
		dstVersion = string(data)
	}

	dstInstalled := false
	binDir := filepath.Join(h.dataDir, "dst_server", "bin64")
	if entries, err := os.ReadDir(binDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && (e.Name() == "dontstarve_dedicated_server_nullrenderer" ||
				e.Name() == "dontstarve_dedicated_server_nullrenderer_x64") {
				dstInstalled = true
				break
			}
		}
	}

	// Show Install/Update DST button if brew is available (can auto-install DepotDownloader)
	// or if DepotDownloader is already installed, or if DST is already host-mounted
	_, brewErr := exec.LookPath("brew")
	_, depotErr := exec.LookPath("DepotDownloader")
	needsManualUpdate := dstInstalled || depotErr == nil || brewErr == nil

	writeJSON(w, http.StatusOK, map[string]any{
		"imageExists":       exists,
		"dstInstalled":      dstInstalled,
		"dstVersion":        dstVersion,
		"needsManualUpdate": needsManualUpdate,
	})
}

func (h *Handler) BuildImage(w http.ResponseWriter, r *http.Request) {
	dockerDir := ""
	candidates := []string{
		filepath.Join(h.dataDir, "..", "docker"),
		filepath.Join(h.dataDir, "docker"),
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "..", "..", "..", "docker"))
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "docker"))
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "Dockerfile.dst")); err == nil {
			dockerDir = c
			break
		}
	}
	if dockerDir == "" {
		writeError(w, http.StatusBadRequest, "Docker files not found. Run 'docker build' manually from project directory.")
		return
	}

	cmd := exec.CommandContext(r.Context(), "docker", "build", "-f",
		filepath.Join(dockerDir, "Dockerfile.dst"), "-t", "dst-server:latest", dockerDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":  err.Error(),
			"output": string(output),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "built",
		"output": string(output),
	})
}

func (h *Handler) UpdateDST(w http.ResponseWriter, r *http.Request) {
	depotPath, err := exec.LookPath("DepotDownloader")
	if err != nil {
		// Auto-install DepotDownloader via Homebrew
		brewPath, brewErr := exec.LookPath("brew")
		if brewErr != nil {
			writeError(w, http.StatusBadRequest, "Homebrew not found. Install from https://brew.sh")
			return
		}

		log.Println("DepotDownloader not found, installing via Homebrew...")

		// brew tap + install
		tapCmd := exec.CommandContext(r.Context(), brewPath, "tap", "steamre/tools")
		tapOut, tapErr := tapCmd.CombinedOutput()
		log.Printf("brew tap: %s", string(tapOut))

		installCmd := exec.CommandContext(r.Context(), brewPath, "install", "depotdownloader")
		installOut, installErr := installCmd.CombinedOutput()
		log.Printf("brew install: %s", string(installOut))

		if tapErr != nil || installErr != nil {
			errMsg := fmt.Sprintf("Failed to install DepotDownloader.\ntap: %v\ninstall: %v\n%s\n%s",
				tapErr, installErr, string(tapOut), string(installOut))
			writeError(w, http.StatusInternalServerError, errMsg)
			return
		}

		// Find the newly installed binary
		depotPath, err = exec.LookPath("DepotDownloader")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DepotDownloader installed but not found in PATH")
			return
		}
		log.Printf("DepotDownloader installed at: %s", depotPath)
	}

	beta := r.URL.Query().Get("beta")

	dstDir := filepath.Join(h.dataDir, "dst_server")
	os.MkdirAll(dstDir, 0755)

	logPath := filepath.Join(h.dataDir, "install-dst.log")
	log.Printf("Updating DST server via DepotDownloader (beta=%s, log: %s)", beta, logPath)

	args := []string{"-app", "343050", "-os", "linux", "-dir", dstDir}
	if beta != "" {
		args = append(args, "-beta", beta)
	}
	cmd := exec.CommandContext(r.Context(), depotPath, args...)
	output, err := cmd.CombinedOutput()

	// Write log file
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("=== DST Install/Update: %s ===\nCommand: DepotDownloader -app 343050 -os linux -dir %s\n\n%s\n", timestamp, dstDir, string(output))
	if err != nil {
		logEntry += fmt.Sprintf("\nERROR: %s\n", err.Error())
	} else {
		logEntry += "\nSUCCESS\n"
	}
	logFile, logErr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logErr == nil {
		logFile.WriteString(logEntry)
		logFile.Close()
	}

	// Fix permissions
	binDir := filepath.Join(dstDir, "bin64")
	if entries, dirErr := os.ReadDir(binDir); dirErr == nil {
		for _, e := range entries {
			if !e.IsDir() {
				os.Chmod(filepath.Join(binDir, e.Name()), 0755)
			}
		}
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":  err.Error(),
			"output": string(output),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "updated",
		"output": string(output),
	})
}
