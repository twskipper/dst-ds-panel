package docker

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
)

func (m *Manager) StartShard(ctx context.Context, clusterDirName, shardName string) (string, error) {
	containerName := fmt.Sprintf("dst-%s-%s", strings.ToLower(clusterDirName), strings.ToLower(shardName))

	// Remove existing container with same name if any
	_ = m.removeShard(ctx, containerName)

	binds := []string{
		fmt.Sprintf("%s/clusters/%s:/root/.klei/DoNotStarveTogether/%s", m.dataDir, clusterDirName, clusterDirName),
	}
	// Mount DST server from host if available (macOS mode), otherwise container has it built-in (Linux mode)
	dstServerDir := fmt.Sprintf("%s/dst_server", m.dataDir)
	if _, err := os.Stat(filepath.Join(dstServerDir, "bin64")); err == nil {
		binds = append(binds, fmt.Sprintf("%s:/opt/dst_server", dstServerDir))
	}

	hostConfig := &container.HostConfig{
		Binds: binds,
		NetworkMode: "host",
		RestartPolicy: container.RestartPolicy{
			Name:              container.RestartPolicyOnFailure,
			MaximumRetryCount: 3,
		},
	}

	cfg := &container.Config{
		Image:       m.imageName,
		AttachStdin: true,
		OpenStdin:   true,
		Env: []string{
			fmt.Sprintf("CLUSTER_NAME=%s", clusterDirName),
			fmt.Sprintf("SHARD=%s", shardName),
		},
		Labels: map[string]string{
			"managed-by": ContainerLabel,
			"cluster":    clusterDirName,
			"shard":      shardName,
		},
	}

	resp, err := m.cli.ContainerCreate(ctx, cfg, hostConfig, nil, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}

	if err := m.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("start container: %w", err)
	}

	return resp.ID, nil
}

func (m *Manager) StopShard(ctx context.Context, containerID string) error {
	timeout := 30
	stopOpts := container.StopOptions{Timeout: &timeout}
	if err := m.cli.ContainerStop(ctx, containerID, stopOpts); err != nil {
		return err
	}
	return m.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{})
}

func (m *Manager) removeShard(ctx context.Context, containerName string) error {
	f := filters.NewArgs(filters.Arg("name", containerName))
	containers, err := m.cli.ContainerList(ctx, container.ListOptions{All: true, Filters: f})
	if err != nil {
		return err
	}
	for _, c := range containers {
		timeout := 10
		_ = m.cli.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeout})
		_ = m.cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true})
	}
	return nil
}

// ListRunningShards returns a map of "cluster-shard" -> containerID for all running DST containers
func (m *Manager) ListRunningShards(ctx context.Context) (map[string]string, error) {
	f := filters.NewArgs(filters.Arg("label", "managed-by="+ContainerLabel))
	containers, err := m.cli.ContainerList(ctx, container.ListOptions{Filters: f})
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, c := range containers {
		cluster := c.Labels["cluster"]
		shard := c.Labels["shard"]
		if cluster != "" && shard != "" {
			result[cluster+"/"+shard] = c.ID
		}
	}
	return result, nil
}

func (m *Manager) ExecCommand(ctx context.Context, containerID, command string) error {
	// Use printf instead of echo to avoid shell escaping issues with quotes
	// printf '%s\n' writes the command literally without interpreting special chars
	execCfg, err := m.cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          []string{"sh", "-c", "printf '%s\\n' " + shellQuote(command) + " > /proc/1/fd/0"},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return fmt.Errorf("exec create: %w", err)
	}

	if err := m.cli.ContainerExecStart(ctx, execCfg.ID, container.ExecStartOptions{}); err != nil {
		return fmt.Errorf("exec start: %w", err)
	}

	return nil
}

// shellQuote wraps a string in single quotes, escaping any embedded single quotes
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func (m *Manager) StreamLogs(ctx context.Context, containerID string) (io.ReadCloser, error) {
	return m.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "100",
	})
}

func (m *Manager) StreamLogsLines(ctx context.Context, containerID string) (<-chan string, error) {
	reader, err := m.StreamLogs(ctx, containerID)
	if err != nil {
		return nil, err
	}

	ch := make(chan string, 64)
	go func() {
		defer close(ch)
		defer reader.Close()

		// Docker multiplexed log stream has 8-byte headers per frame.
		// Use stdcopy to demux stdout/stderr into a clean stream.
		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			stdcopy.StdCopy(pw, pw, reader)
		}()

		scanner := bufio.NewScanner(pr)
		// Increase buffer for long log lines
		scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
		for scanner.Scan() {
			line := scanner.Text()
			// Split on embedded newlines (DST sometimes bundles multiple lines)
			for _, part := range splitLines(line) {
				if part == "" {
					continue
				}
				select {
				case ch <- part:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch, nil
}

func splitLines(s string) []string {
	var lines []string
	buf := bytes.NewBufferString(s)
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(lines) == 0 {
		return []string{s}
	}
	return lines
}
