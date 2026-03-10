package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
)

func (m *Manager) ImageExists(ctx context.Context) (bool, error) {
	f := filters.NewArgs(filters.Arg("reference", m.imageName))
	images, err := m.cli.ImageList(ctx, image.ListOptions{Filters: f})
	if err != nil {
		return false, err
	}
	return len(images) > 0, nil
}

func (m *Manager) BuildImage(ctx context.Context, dockerfileDir string) (io.ReadCloser, error) {
	// For image building, we use the CLI approach via exec since the Docker SDK
	// build API requires a tar context which is complex. The handler will use
	// exec.Command("docker", "build", ...) instead.
	return nil, nil
}
