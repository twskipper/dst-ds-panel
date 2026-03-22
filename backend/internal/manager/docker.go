package manager

import (
	"context"

	"dst-ds-panel/internal/docker"
	"dst-ds-panel/internal/model"
)

// DockerManager wraps docker.Manager to implement ShardManager.
type DockerManager struct {
	mgr *docker.Manager
}

func NewDockerManager(dataDir, imageName, platform string) (*DockerManager, error) {
	mgr, err := docker.NewManager(dataDir, imageName, platform)
	if err != nil {
		return nil, err
	}
	return &DockerManager{mgr: mgr}, nil
}

func (d *DockerManager) StartShard(ctx context.Context, clusterDirName, shardName string) (string, error) {
	return d.mgr.StartShard(ctx, clusterDirName, shardName)
}

func (d *DockerManager) StopShard(ctx context.Context, id string) error {
	return d.mgr.StopShard(ctx, id)
}

func (d *DockerManager) ListRunningShards(ctx context.Context) (map[string]string, error) {
	return d.mgr.ListRunningShards(ctx)
}

func (d *DockerManager) ExecCommand(ctx context.Context, id, command string) error {
	return d.mgr.ExecCommand(ctx, id, command)
}

func (d *DockerManager) StreamLogsLines(ctx context.Context, id string) (<-chan string, error) {
	return d.mgr.StreamLogsLines(ctx, id)
}

func (d *DockerManager) StreamStats(ctx context.Context, id string) (<-chan model.ContainerStats, error) {
	return d.mgr.StreamStats(ctx, id)
}

func (d *DockerManager) ImageExists(ctx context.Context) (bool, error) {
	return d.mgr.ImageExists(ctx)
}

func (d *DockerManager) EnsureImage(ctx context.Context) error {
	return d.mgr.EnsureImage(ctx)
}

func (d *DockerManager) Close() error {
	return d.mgr.Close()
}
