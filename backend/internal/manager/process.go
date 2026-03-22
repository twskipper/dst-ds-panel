package manager

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"dst-ds-panel/internal/model"
)

// ProcessManager runs DST server shards as native OS processes (no Docker).
type ProcessManager struct {
	dataDir string
	mu      sync.RWMutex
	procs   map[string]*shardProcess // key: "clusterDirName/shardName"
}

type shardProcess struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	logLines  []string
	logMu     sync.RWMutex
	logSubs   []chan string
	subsMu    sync.Mutex
	cluster   string
	shard     string
	pid       int
	startTime time.Time
}

func NewProcessManager(dataDir string) *ProcessManager {
	return &ProcessManager{
		dataDir: dataDir,
		procs:   make(map[string]*shardProcess),
	}
}

func (m *ProcessManager) findDSTBinary() (string, error) {
	binDir := filepath.Join(m.dataDir, "dst_server", "bin64")

	// Try platform-appropriate binaries
	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{
			"dontstarve_dedicated_server_nullrenderer_x64.exe",
			"dontstarve_dedicated_server_nullrenderer.exe",
		}
	} else {
		candidates = []string{
			"dontstarve_dedicated_server_nullrenderer_x64",
			"dontstarve_dedicated_server_nullrenderer",
		}
	}

	for _, name := range candidates {
		p := filepath.Join(binDir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("DST server binary not found in %s", binDir)
}

func (m *ProcessManager) StartShard(ctx context.Context, clusterDirName, shardName string) (string, error) {
	key := clusterDirName + "/" + shardName

	m.mu.Lock()
	if existing, ok := m.procs[key]; ok {
		if existing.cmd.Process != nil {
			existing.cmd.Process.Kill()
		}
		delete(m.procs, key)
	}
	m.mu.Unlock()

	binPath, err := m.findDSTBinary()
	if err != nil {
		return "", err
	}

	// Set up cluster path for DST
	clusterRoot := filepath.Join(m.dataDir, "clusters")

	// Copy mods setup if provided
	modsSetup := filepath.Join(clusterRoot, clusterDirName, "mods_setup.lua")
	if _, err := os.Stat(modsSetup); err == nil {
		dstModsDir := filepath.Join(m.dataDir, "dst_server", "mods")
		os.MkdirAll(dstModsDir, 0755)
		src, _ := os.ReadFile(modsSetup)
		os.WriteFile(filepath.Join(dstModsDir, "dedicated_server_mods_setup.lua"), src, 0644)
	}

	cmd := exec.Command(binPath,
		"-persistent_storage_root", clusterRoot,
		"-conf_dir", ".",
		"-cluster", clusterDirName,
		"-shard", shardName,
		"-console",
	)
	cmd.Dir = filepath.Dir(binPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start process: %w", err)
	}

	sp := &shardProcess{
		cmd:       cmd,
		stdin:     stdin,
		cluster:   clusterDirName,
		shard:     shardName,
		pid:       cmd.Process.Pid,
		startTime: time.Now(),
	}

	// Read stdout in background, store log lines and broadcast to subscribers
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
		for scanner.Scan() {
			line := scanner.Text()
			sp.logMu.Lock()
			sp.logLines = append(sp.logLines, line)
			// Keep last 10000 lines
			if len(sp.logLines) > 10000 {
				sp.logLines = sp.logLines[len(sp.logLines)-5000:]
			}
			sp.logMu.Unlock()

			sp.subsMu.Lock()
			for _, ch := range sp.logSubs {
				select {
				case ch <- line:
				default:
				}
			}
			sp.subsMu.Unlock()
		}
	}()

	// Wait for process exit in background
	go func() {
		cmd.Wait()
		m.mu.Lock()
		delete(m.procs, key)
		m.mu.Unlock()
		log.Printf("Process exited: %s/%s (pid %d)", clusterDirName, shardName, sp.pid)
	}()

	id := strconv.Itoa(sp.pid)
	m.mu.Lock()
	m.procs[key] = sp
	m.mu.Unlock()

	log.Printf("Started shard %s/%s (pid %d)", clusterDirName, shardName, sp.pid)
	return id, nil
}

func (m *ProcessManager) StopShard(ctx context.Context, id string) error {
	m.mu.RLock()
	var sp *shardProcess
	for _, p := range m.procs {
		if strconv.Itoa(p.pid) == id {
			sp = p
			break
		}
	}
	m.mu.RUnlock()

	if sp == nil {
		return nil // already stopped
	}

	// Try graceful shutdown via console command
	sp.stdin.Write([]byte("c_shutdown(true)\n"))

	// Wait up to 15 seconds for graceful exit
	done := make(chan struct{})
	go func() {
		sp.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(15 * time.Second):
		log.Printf("Force killing shard pid %s", id)
		if sp.cmd.Process != nil {
			sp.cmd.Process.Kill()
		}
		return nil
	}
}

func (m *ProcessManager) ListRunningShards(ctx context.Context) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]string)
	for key, sp := range m.procs {
		if sp.cmd.Process != nil && sp.cmd.ProcessState == nil {
			result[key] = strconv.Itoa(sp.pid)
		}
	}
	return result, nil
}

func (m *ProcessManager) ExecCommand(ctx context.Context, id, command string) error {
	m.mu.RLock()
	var sp *shardProcess
	for _, p := range m.procs {
		if strconv.Itoa(p.pid) == id {
			sp = p
			break
		}
	}
	m.mu.RUnlock()

	if sp == nil {
		return fmt.Errorf("process %s not found", id)
	}

	_, err := sp.stdin.Write([]byte(command + "\n"))
	return err
}

func (m *ProcessManager) StreamLogsLines(ctx context.Context, id string) (<-chan string, error) {
	m.mu.RLock()
	var sp *shardProcess
	for _, p := range m.procs {
		if strconv.Itoa(p.pid) == id {
			sp = p
			break
		}
	}
	m.mu.RUnlock()

	if sp == nil {
		return nil, fmt.Errorf("process %s not found", id)
	}

	ch := make(chan string, 64)

	// Send last 100 lines of history
	sp.logMu.RLock()
	start := len(sp.logLines) - 100
	if start < 0 {
		start = 0
	}
	history := make([]string, len(sp.logLines[start:]))
	copy(history, sp.logLines[start:])
	sp.logMu.RUnlock()

	go func() {
		defer func() {
			sp.subsMu.Lock()
			for i, sub := range sp.logSubs {
				if sub == ch {
					sp.logSubs = append(sp.logSubs[:i], sp.logSubs[i+1:]...)
					break
				}
			}
			sp.subsMu.Unlock()
		}()

		// Send history
		for _, line := range history {
			select {
			case ch <- line:
			case <-ctx.Done():
				close(ch)
				return
			}
		}

		// Subscribe for new lines
		sp.subsMu.Lock()
		sp.logSubs = append(sp.logSubs, ch)
		sp.subsMu.Unlock()

		// Wait until context cancelled
		<-ctx.Done()
		close(ch)
	}()

	return ch, nil
}

func (m *ProcessManager) StreamStats(ctx context.Context, id string) (<-chan model.ContainerStats, error) {
	m.mu.RLock()
	var sp *shardProcess
	for _, p := range m.procs {
		if strconv.Itoa(p.pid) == id {
			sp = p
			break
		}
	}
	m.mu.RUnlock()

	if sp == nil {
		return nil, fmt.Errorf("process %s not found", id)
	}

	ch := make(chan model.ContainerStats, 16)
	go func() {
		defer close(ch)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				stats := model.ContainerStats{
					Timestamp: time.Now().Unix(),
				}
				// Basic stats from /proc on Linux or equivalent
				// For simplicity, report elapsed time and basic info
				if sp.cmd.ProcessState != nil {
					return // process exited
				}
				stats.CPUPercent = 0  // Platform-specific; placeholder
				stats.MemUsageMB = 0 // Platform-specific; placeholder
				stats.MemLimitMB = 0

				select {
				case ch <- stats:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func (m *ProcessManager) ImageExists(ctx context.Context) (bool, error) {
	// Native mode doesn't use Docker images
	return true, nil
}

func (m *ProcessManager) EnsureImage(ctx context.Context) error {
	// Native mode doesn't need Docker images
	return nil
}

func (m *ProcessManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, sp := range m.procs {
		log.Printf("Shutting down shard %s (pid %d)", key, sp.pid)
		sp.stdin.Write([]byte("c_shutdown(true)\n"))
		done := make(chan struct{})
		go func() {
			sp.cmd.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			if sp.cmd.Process != nil {
				sp.cmd.Process.Kill()
			}
		}
	}
	m.procs = make(map[string]*shardProcess)
	return nil
}

// findProcess finds a shardProcess by PID string
func (m *ProcessManager) findProcess(id string) *shardProcess {
	for _, p := range m.procs {
		if strconv.Itoa(p.pid) == id {
			return p
		}
	}
	return nil
}

// DSTBinaryDir returns the path where DST binaries are expected
func (m *ProcessManager) DSTBinaryDir() string {
	return filepath.Join(m.dataDir, "dst_server", "bin64")
}

// DSTInstalled checks if the DST server binary exists
func (m *ProcessManager) DSTInstalled() bool {
	_, err := m.findDSTBinary()
	return err == nil
}

