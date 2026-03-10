package docker

import (
	"github.com/docker/docker/client"
)

const (
	ContainerLabel = "dst-ds-panel"
)

type Manager struct {
	cli       *client.Client
	imageName string
	platform  string
	dataDir   string
}

func NewManager(dataDir, imageName, platform string) (*Manager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Manager{
		cli:       cli,
		imageName: imageName,
		platform:  platform,
		dataDir:   dataDir,
	}, nil
}

func (m *Manager) Close() error {
	return m.cli.Close()
}
