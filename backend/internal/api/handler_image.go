package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func (h *Handler) ImageStatus(w http.ResponseWriter, r *http.Request) {
	exists, err := h.shardMgr.ImageExists(r.Context())
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
				e.Name() == "dontstarve_dedicated_server_nullrenderer_x64" ||
				e.Name() == "dontstarve_dedicated_server_nullrenderer_x64.exe") {
				dstInstalled = true
				break
			}
		}
	}

	dstBranch := ""
	branchFile := filepath.Join(h.dataDir, "dst_server", "branch.txt")
	if data, err := os.ReadFile(branchFile); err == nil {
		dstBranch = string(data)
	}

	// Determine if we can show the Install/Update DST button
	needsManualUpdate := false
	if h.mode == "native" {
		// Native mode: always show update button (DepotDownloader can be auto-downloaded)
		needsManualUpdate = true
	} else {
		// Docker mode: show if brew or DepotDownloader available
		_, brewErr := exec.LookPath("brew")
		_, depotErr := exec.LookPath("DepotDownloader")
		needsManualUpdate = dstInstalled || depotErr == nil || brewErr == nil
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"imageExists":       exists,
		"dstInstalled":      dstInstalled,
		"dstVersion":        dstVersion,
		"dstBranch":         dstBranch,
		"needsManualUpdate": needsManualUpdate,
		"mode":              h.mode,
	})
}

func (h *Handler) BuildImage(w http.ResponseWriter, r *http.Request) {
	if h.mode == "native" {
		writeError(w, http.StatusBadRequest, "Docker image build not available in native mode")
		return
	}

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

func (h *Handler) findDepotDownloader(ctx context.Context) (string, error) {
	// Check PATH first
	if p, err := exec.LookPath("DepotDownloader"); err == nil {
		return p, nil
	}

	// Check our tools directory
	toolsDir := filepath.Join(h.dataDir, "tools")
	localBin := filepath.Join(toolsDir, "DepotDownloader")
	if runtime.GOOS == "windows" {
		localBin = filepath.Join(toolsDir, "DepotDownloader.exe")
	}
	if _, err := os.Stat(localBin); err == nil {
		return localBin, nil
	}

	// Try to auto-install
	if runtime.GOOS == "darwin" {
		// macOS: use Homebrew
		brewPath, err := exec.LookPath("brew")
		if err != nil {
			return "", fmt.Errorf("DepotDownloader not found. Install Homebrew from https://brew.sh")
		}

		log.Println("Installing DepotDownloader via Homebrew...")
		tapCmd := exec.CommandContext(ctx, brewPath, "tap", "steamre/tools")
		tapOut, _ := tapCmd.CombinedOutput()
		log.Printf("brew tap: %s", string(tapOut))

		installCmd := exec.CommandContext(ctx, brewPath, "install", "depotdownloader")
		installOut, installErr := installCmd.CombinedOutput()
		log.Printf("brew install: %s", string(installOut))

		if installErr != nil {
			return "", fmt.Errorf("failed to install DepotDownloader: %v\n%s", installErr, string(installOut))
		}

		if p, err := exec.LookPath("DepotDownloader"); err == nil {
			return p, nil
		}
		return "", fmt.Errorf("DepotDownloader installed but not found in PATH")
	}

	// Windows/Linux: auto-download standalone release from GitHub
	return h.downloadDepotDownloader(ctx, toolsDir)
}

func (h *Handler) downloadDepotDownloader(ctx context.Context, toolsDir string) (string, error) {
	// Determine platform suffix for GitHub release
	var assetName string
	switch runtime.GOOS {
	case "windows":
		assetName = "DepotDownloader-windows-x64.zip"
	case "linux":
		assetName = "DepotDownloader-linux-x64.zip"
	default:
		return "", fmt.Errorf("unsupported platform for auto-download: %s", runtime.GOOS)
	}

	os.MkdirAll(toolsDir, 0755)

	// Use GitHub API to get latest release download URL
	releaseURL := "https://github.com/SteamRE/DepotDownloader/releases/latest/download/" + assetName
	log.Printf("Downloading DepotDownloader from %s", releaseURL)

	zipPath := filepath.Join(toolsDir, assetName)

	// Download the zip
	dlCmd := exec.CommandContext(ctx, "curl", "-L", "-o", zipPath, releaseURL)
	dlOut, dlErr := dlCmd.CombinedOutput()
	if dlErr != nil {
		return "", fmt.Errorf("failed to download DepotDownloader: %v\n%s", dlErr, string(dlOut))
	}

	// Extract the zip
	if runtime.GOOS == "windows" {
		// Use PowerShell to extract on Windows
		psCmd := exec.CommandContext(ctx, "powershell", "-Command",
			fmt.Sprintf("Expand-Archive -Force -Path '%s' -DestinationPath '%s'", zipPath, toolsDir))
		psOut, psErr := psCmd.CombinedOutput()
		if psErr != nil {
			return "", fmt.Errorf("failed to extract DepotDownloader: %v\n%s", psErr, string(psOut))
		}
	} else {
		// Use unzip on Linux
		uzCmd := exec.CommandContext(ctx, "unzip", "-o", zipPath, "-d", toolsDir)
		uzOut, uzErr := uzCmd.CombinedOutput()
		if uzErr != nil {
			return "", fmt.Errorf("failed to extract DepotDownloader: %v\n%s", uzErr, string(uzOut))
		}
	}

	// Clean up zip
	os.Remove(zipPath)

	// Find the binary
	binName := "DepotDownloader"
	if runtime.GOOS == "windows" {
		binName = "DepotDownloader.exe"
	}
	binPath := filepath.Join(toolsDir, binName)
	if _, err := os.Stat(binPath); err != nil {
		return "", fmt.Errorf("DepotDownloader extracted but binary not found at %s", binPath)
	}

	// Make executable (Linux)
	if runtime.GOOS != "windows" {
		os.Chmod(binPath, 0755)
	}

	log.Printf("DepotDownloader installed at: %s", binPath)
	return binPath, nil
}

func (h *Handler) UpdateDST(w http.ResponseWriter, r *http.Request) {
	depotPath, err := h.findDepotDownloader(r.Context())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	branch := r.URL.Query().Get("branch")

	// Determine target OS for DepotDownloader
	targetOS := "linux"
	if h.mode == "native" && runtime.GOOS == "windows" {
		targetOS = "windows"
	}

	dstDir := filepath.Join(h.dataDir, "dst_server")
	os.MkdirAll(dstDir, 0755)

	logPath := filepath.Join(h.dataDir, "install-dst.log")
	log.Printf("Updating DST server via DepotDownloader (os=%s, branch=%s, log: %s)", targetOS, branch, logPath)

	args := []string{"-app", "343050", "-os", targetOS, "-dir", dstDir}
	if branch != "" {
		args = append(args, "-branch", branch)
	}
	cmd := exec.CommandContext(r.Context(), depotPath, args...)
	output, err := cmd.CombinedOutput()

	// Write log file
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("=== DST Install/Update: %s ===\nCommand: DepotDownloader -app 343050 -os %s -dir %s\n\n%s\n", timestamp, targetOS, dstDir, string(output))
	if err != nil {
		logEntry += fmt.Sprintf("\nERROR: %s\n", err.Error())
	} else {
		logEntry += "\nSUCCESS\n"
		branchLabel := "public"
		if branch != "" {
			branchLabel = branch
		}
		os.WriteFile(filepath.Join(dstDir, "branch.txt"), []byte(branchLabel), 0644)
	}
	logFile, logErr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logErr == nil {
		logFile.WriteString(logEntry)
		logFile.Close()
	}

	// Fix permissions (Linux/macOS)
	if runtime.GOOS != "windows" {
		binDir := filepath.Join(dstDir, "bin64")
		if entries, dirErr := os.ReadDir(binDir); dirErr == nil {
			for _, e := range entries {
				if !e.IsDir() {
					os.Chmod(filepath.Join(binDir, e.Name()), 0755)
				}
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
