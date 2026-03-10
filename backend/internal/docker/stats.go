package docker

import (
	"context"
	"encoding/json"
	"time"

	"dst-ds-panel/internal/model"

	"github.com/docker/docker/api/types/container"
)

func (m *Manager) StreamStats(ctx context.Context, containerID string) (<-chan model.ContainerStats, error) {
	resp, err := m.cli.ContainerStats(ctx, containerID, true)
	if err != nil {
		return nil, err
	}

	ch := make(chan model.ContainerStats, 16)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		for {
			var stats container.StatsResponse
			if err := decoder.Decode(&stats); err != nil {
				return
			}

			cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
			systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
			cpuPercent := 0.0
			if systemDelta > 0 && cpuDelta > 0 {
				cpuPercent = (cpuDelta / systemDelta) * float64(stats.CPUStats.OnlineCPUs) * 100.0
			}

			memUsage := float64(stats.MemoryStats.Usage) / 1024 / 1024
			memLimit := float64(stats.MemoryStats.Limit) / 1024 / 1024

			select {
			case ch <- model.ContainerStats{
				CPUPercent: cpuPercent,
				MemUsageMB: memUsage,
				MemLimitMB: memLimit,
				Timestamp:  time.Now().Unix(),
			}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}
