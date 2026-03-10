package docker

import (
	"context"
	"io"
	"log"

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

func (m *Manager) EnsureImage(ctx context.Context) error {
	exists, err := m.ImageExists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	log.Printf("Image %s not found, pulling...", m.imageName)
	reader, err := m.cli.ImagePull(ctx, m.imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	// Read to completion to ensure pull finishes
	io.Copy(io.Discard, reader)
	log.Printf("Image %s pulled successfully", m.imageName)
	return nil
}

func (m *Manager) BuildImage(ctx context.Context, dockerfileDir string) (io.ReadCloser, error) {
	return nil, nil
}
