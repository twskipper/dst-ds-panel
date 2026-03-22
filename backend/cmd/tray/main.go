package main

import (
	_ "embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/getlantern/systray"
)

//go:embed assets/icon.png
var iconData []byte

var (
	serverProcess *os.Process
	serverMu      sync.Mutex
	serverRunning bool
	serverLogFile *os.File
	appDir        string
)

func main() {
	// Determine app directory
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	appDir = filepath.Dir(exe)

	// Determine data directory based on platform
	dataHome := getDataDir()
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
	if runtime.GOOS == "darwin" {
		currentPath := os.Getenv("PATH")
		os.Setenv("PATH", "/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin:"+currentPath)
	}

	os.Chdir(dataHome)
	log.Printf("Data directory: %s", dataHome)
	log.Printf("Binary directory: %s", appDir)

	systray.Run(onReady, onExit)
}

func getDataDir() string {
	homeDir, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "DST DS Panel")
	case "windows":
		// Portable: store data next to the exe
		return appDir
	default:
		// Linux: XDG_DATA_HOME or ~/.local/share
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "dst-ds-panel")
		}
		return filepath.Join(homeDir, ".local", "share", "dst-ds-panel")
	}
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("DST")
	systray.SetTooltip("DST DS Panel")

	// Menu items
	mStatus := systray.AddMenuItem("Checking status...", "")
	mStatus.Disable()

	systray.AddSeparator()

	mStartStop := systray.AddMenuItem("Start Server", "Start the DST DS Panel backend")
	mOpen := systray.AddMenuItem("Open Panel", "Open in browser")
	mOpen.Disable()
	mOpenData := systray.AddMenuItem("Open Data Folder", "Open data directory")

	systray.AddSeparator()

	// Platform-specific dependency items
	var depItems []*systray.MenuItem
	if runtime.GOOS == "darwin" {
		mDocker := systray.AddMenuItem("Docker: checking...", "")
		mDocker.Disable()
		mBrew := systray.AddMenuItem("Homebrew: checking...", "")
		mBrew.Disable()
		depItems = append(depItems, mDocker, mBrew)
	}
	// Windows native mode: no external dependencies needed

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit DST DS Panel")

	// Check dependencies in background
	go func() {
		if runtime.GOOS == "darwin" && len(depItems) >= 2 {
			dockerOk := findBinary("docker") != ""
			if dockerOk {
				depItems[0].SetTitle("Docker: ✓ installed")
			} else {
				depItems[0].SetTitle("Docker: ✗ not found")
			}

			brewOk := findBinary("brew") != ""
			if brewOk {
				depItems[1].SetTitle("Homebrew: ✓ installed")
			} else {
				depItems[1].SetTitle("Homebrew: ✗ not found")
			}

			if dockerOk && brewOk {
				mStatus.SetTitle("Ready")
			} else {
				mStatus.SetTitle("Missing: Docker or Homebrew")
			}
		} else {
			// Windows/Linux native mode
			mStatus.SetTitle("Ready")
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
				openFolder()

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
	serverBin := findServerBinary()
	if serverBin == "" {
		log.Println("Server binary not found")
		return false
	}

	cwd, _ := os.Getwd()
	cmd := exec.Command(serverBin)
	cmd.Dir = cwd

	// Log server output to file
	logPath := filepath.Join(cwd, "dst-panel-server.log")
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = lf
		cmd.Stderr = lf
		serverLogFile = lf
	}

	if runtime.GOOS == "darwin" {
		cmd.Env = append(os.Environ(),
			"PATH=/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin",
		)
	}

	// Hide console window on Windows
	hideConsoleWindow(cmd)

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start server: %v", err)
		if serverLogFile != nil {
			serverLogFile.Close()
			serverLogFile = nil
		}
		return false
	}

	serverProcess = cmd.Process
	serverRunning = true
	log.Printf("Server started (PID: %d), log: %s", serverProcess.Pid, logPath)

	go func() {
		cmd.Wait()
		serverMu.Lock()
		serverRunning = false
		serverProcess = nil
		if serverLogFile != nil {
			serverLogFile.Close()
			serverLogFile = nil
		}
		serverMu.Unlock()
		log.Println("Server process exited")
	}()

	time.Sleep(500 * time.Millisecond)
	return true
}

func stopServer() {
	if serverProcess != nil {
		log.Println("Stopping server...")
		serverProcess.Signal(os.Interrupt)

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
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	candidates := []string{
		filepath.Join(appDir, "dst-ds-panel"+ext),
		filepath.Join(appDir, "..", "dst-ds-panel"+ext),
		filepath.Join(appDir, "..", "backend", "dst-ds-panel"+ext),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	if p, err := exec.LookPath("dst-ds-panel" + ext); err == nil {
		return p
	}
	return ""
}

func findBinary(name string) string {
	if runtime.GOOS == "darwin" {
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
	}
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
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Run()
}

func openFolder() {
	cwd, _ := os.Getwd()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", cwd)
	case "windows":
		cmd = exec.Command("explorer", cwd)
	default:
		cmd = exec.Command("xdg-open", cwd)
	}
	cmd.Run()
}
