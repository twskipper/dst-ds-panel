package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/getlantern/systray"
)

var (
	serverProcess *os.Process
	serverMu      sync.Mutex
	serverRunning bool
	appDir        string
)

func main() {
	// Determine app directory
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	appDir = filepath.Dir(exe)

	// Use ~/Library/Application Support/DST DS Panel as data directory
	// This is the standard macOS location for app data
	homeDir, _ := os.UserHomeDir()
	dataHome := filepath.Join(homeDir, "Library", "Application Support", "DST DS Panel")
	os.MkdirAll(dataHome, 0755)

	// Copy config.example.json if config.json doesn't exist
	configPath := filepath.Join(dataHome, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		exampleConfig := filepath.Join(appDir, "config.example.json")
		if data, err := os.ReadFile(exampleConfig); err == nil {
			os.WriteFile(configPath, data, 0644)
			log.Printf("Created default config at: %s", configPath)
		}
	}

	// Fix PATH for macOS GUI apps (they don't inherit shell PATH)
	currentPath := os.Getenv("PATH")
	os.Setenv("PATH", "/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin:"+currentPath)

	os.Chdir(dataHome)
	log.Printf("Data directory: %s", dataHome)
	log.Printf("Binary directory: %s", appDir)

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("DST")
	systray.SetTooltip("DST DS Panel")

	// Menu items
	mStatus := systray.AddMenuItem("Checking status...", "")
	mStatus.Disable()

	systray.AddSeparator()

	mStartStop := systray.AddMenuItem("Start Server", "Start the DST DS Panel backend")
	mOpen := systray.AddMenuItem("Open Panel", "Open in browser")
	mOpen.Disable()
	mOpenData := systray.AddMenuItem("Open Data Folder", "Open data directory in Finder")

	systray.AddSeparator()

	mDocker := systray.AddMenuItem("Docker: checking...", "")
	mDocker.Disable()
	mBrew := systray.AddMenuItem("Homebrew: checking...", "")
	mBrew.Disable()
	mDepot := systray.AddMenuItem("DepotDownloader: checking...", "")
	mDepot.Disable()

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit DST DS Panel")

	// Check dependencies in background
	go func() {
		dockerOk := findBinary("docker") != ""
		if dockerOk {
			mDocker.SetTitle("Docker: ✓ installed")
		} else {
			mDocker.SetTitle("Docker: ✗ not found")
		}

		brewOk := findBinary("brew") != ""
		if brewOk {
			mBrew.SetTitle("Homebrew: ✓ installed")
		} else {
			mBrew.SetTitle("Homebrew: ✗ not found")
		}

		depotOk := findBinary("DepotDownloader") != ""
		if depotOk {
			mDepot.SetTitle("DepotDownloader: ✓ installed")
		} else {
			mDepot.SetTitle("DepotDownloader: ✗ not found")
		}

		if dockerOk {
			mStatus.SetTitle("Ready — Docker available")
		} else {
			mStatus.SetTitle("Docker required — install Docker first")
		}
	}()

	// Event loop
	go func() {
		for {
			select {
			case <-mStartStop.ClickedCh:
				serverMu.Lock()
				if serverRunning {
					stopServer()
					mStartStop.SetTitle("Start Server")
					mOpen.Disable()
					mStatus.SetTitle("Server stopped")
				} else {
					if startServer() {
						mStartStop.SetTitle("Stop Server")
						mOpen.Enable()
						mStatus.SetTitle("Server running on :8080")
					}
				}
				serverMu.Unlock()

			case <-mOpen.ClickedCh:
				openBrowser("http://localhost:8080")

			case <-mOpenData.ClickedCh:
				cwd, _ := os.Getwd()
				exec.Command("open", cwd).Run()

			case <-mQuit.ClickedCh:
				serverMu.Lock()
				if serverRunning {
					stopServer()
				}
				serverMu.Unlock()
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	serverMu.Lock()
	defer serverMu.Unlock()
	if serverRunning {
		stopServer()
	}
}

func startServer() bool {
	// Find the server binary
	serverBin := findServerBinary()
	if serverBin == "" {
		log.Println("Server binary not found")
		return false
	}

	// Run server in the data directory (~/Library/Application Support/DST DS Panel)
	cwd, _ := os.Getwd()
	cmd := exec.Command(serverBin)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Ensure server process has full PATH (macOS GUI apps have limited PATH)
	cmd.Env = append(os.Environ(),
		"PATH=/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin",
	)

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start server: %v", err)
		return false
	}

	serverProcess = cmd.Process
	serverRunning = true
	log.Printf("Server started (PID: %d)", serverProcess.Pid)

	// Monitor process in background
	go func() {
		cmd.Wait()
		serverMu.Lock()
		serverRunning = false
		serverProcess = nil
		serverMu.Unlock()
		log.Println("Server process exited")
	}()

	// Wait a moment for server to start
	time.Sleep(500 * time.Millisecond)
	return true
}

func stopServer() {
	if serverProcess != nil {
		log.Println("Stopping server...")
		serverProcess.Signal(os.Interrupt)

		// Wait up to 5 seconds for graceful shutdown
		done := make(chan struct{})
		go func() {
			serverProcess.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			serverProcess.Kill()
		}

		serverRunning = false
		serverProcess = nil
		log.Println("Server stopped")
	}
}

func findServerBinary() string {
	// Look for server binary in common locations
	candidates := []string{
		filepath.Join(appDir, "dst-ds-panel"),
		filepath.Join(appDir, "..", "dst-ds-panel"),
		filepath.Join(appDir, "..", "backend", "dst-ds-panel"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	// Try PATH
	if p, err := exec.LookPath("dst-ds-panel"); err == nil {
		return p
	}
	return ""
}

func checkCommand(name string, args ...string) bool {
	// macOS GUI apps have limited PATH, add common paths
	fullPath := findBinary(name)
	if fullPath == "" {
		return false
	}
	cmd := exec.Command(fullPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil
}

func findBinary(name string) string {
	// Check common macOS paths that GUI apps don't have in PATH
	searchPaths := []string{
		"/usr/local/bin",
		"/opt/homebrew/bin",
		"/usr/bin",
		"/bin",
		"/usr/sbin",
		"/Applications/OrbStack.app/Contents/MacOS",
	}
	for _, dir := range searchPaths {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// Fallback to PATH lookup
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	return ""
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		cmd = exec.Command("open", url)
	}
	cmd.Run()
}

