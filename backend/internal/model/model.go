package model

import "time"

type ClusterStatus string

const (
	StatusStopped  ClusterStatus = "stopped"
	StatusRunning  ClusterStatus = "running"
	StatusStarting ClusterStatus = "starting"
	StatusError    ClusterStatus = "error"
)

type Cluster struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	DirName   string        `json:"dirName"`
	Status    ClusterStatus `json:"status"`
	Shards    []Shard       `json:"shards"`
	Config    ClusterConfig `json:"config"`
	CreatedAt time.Time     `json:"createdAt"`
}

type Shard struct {
	Name        string        `json:"name"` // "Master" or "Caves"
	ContainerID string        `json:"containerId"`
	Status      ClusterStatus `json:"status"`
	Config      ShardConfig   `json:"config"`
}

type ClusterConfig struct {
	GameMode           string `json:"gameMode"`
	MaxPlayers         int    `json:"maxPlayers"`
	PVP                bool   `json:"pvp"`
	ClusterName        string `json:"clusterName"`
	ClusterDescription string `json:"clusterDescription"`
	ClusterPassword    string `json:"clusterPassword"`
	Token              string `json:"token"`
}

type ShardConfig struct {
	IsMaster   bool `json:"isMaster"`
	ServerPort int  `json:"serverPort"`
	MasterPort int  `json:"masterPort"`
}

type Mod struct {
	WorkshopID string         `json:"workshopId"`
	Name       string         `json:"name"`
	Enabled    bool           `json:"enabled"`
	Config     map[string]any `json:"config"`
}

type ContainerStats struct {
	CPUPercent float64 `json:"cpuPercent"`
	MemUsageMB float64 `json:"memUsageMb"`
	MemLimitMB float64 `json:"memLimitMb"`
	Timestamp  int64   `json:"timestamp"`
}
