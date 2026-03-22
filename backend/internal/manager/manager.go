package manager

import (
	"context"

	"dst-ds-panel/internal/model"
)

// ShardManager abstracts the runtime for DST server shards.
// Docker mode runs shards in containers; native mode runs them as OS processes.
type ShardManager interface {
	StartShard(ctx context.Context, clusterDirName, shardName string) (string, error)
	StopShard(ctx context.Context, id string) error
	ListRunningShards(ctx context.Context) (map[string]string, error)
	ExecCommand(ctx context.Context, id, command string) error
	StreamLogsLines(ctx context.Context, id string) (<-chan string, error)
	StreamStats(ctx context.Context, id string) (<-chan model.ContainerStats, error)
	ImageExists(ctx context.Context) (bool, error)
	EnsureImage(ctx context.Context) error
	Close() error
}
