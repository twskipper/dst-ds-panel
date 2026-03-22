package dst

import (
	"os"
	"path/filepath"
	"strconv"

	"dst-ds-panel/internal/model"

	"gopkg.in/ini.v1"
)

func ReadClusterConfig(clusterDir string) (*model.ClusterConfig, error) {
	cfg, err := ini.Load(filepath.Join(clusterDir, "cluster.ini"))
	if err != nil {
		return nil, err
	}

	gameplay := cfg.Section("GAMEPLAY")
	network := cfg.Section("NETWORK")

	maxPlayers, _ := gameplay.Key("max_players").Int()
	pvp, _ := gameplay.Key("pvp").Bool()

	token := ""
	tokenData, err := os.ReadFile(filepath.Join(clusterDir, "cluster_token.txt"))
	if err == nil {
		token = string(tokenData)
	}

	return &model.ClusterConfig{
		GameMode:           gameplay.Key("game_mode").String(),
		MaxPlayers:         maxPlayers,
		PVP:                pvp,
		ClusterName:        network.Key("cluster_name").String(),
		ClusterDescription: network.Key("cluster_description").String(),
		ClusterPassword:    network.Key("cluster_password").String(),
		Token:              token,
	}, nil
}

func WriteClusterConfig(clusterDir string, config *model.ClusterConfig) error {
	cfg, err := ini.Load(filepath.Join(clusterDir, "cluster.ini"))
	if err != nil {
		cfg = ini.Empty()
	}

	gameplay := cfg.Section("GAMEPLAY")
	gameplay.Key("game_mode").SetValue(config.GameMode)
	gameplay.Key("max_players").SetValue(strconv.Itoa(config.MaxPlayers))
	gameplay.Key("pvp").SetValue(strconv.FormatBool(config.PVP))

	network := cfg.Section("NETWORK")
	network.Key("cluster_name").SetValue(config.ClusterName)
	network.Key("cluster_description").SetValue(config.ClusterDescription)
	network.Key("cluster_password").SetValue(config.ClusterPassword)

	shard := cfg.Section("SHARD")
	if shard.Key("shard_enabled").String() == "" {
		shard.Key("shard_enabled").SetValue("true")
		shard.Key("bind_ip").SetValue("127.0.0.1")
		shard.Key("master_ip").SetValue("127.0.0.1")
		shard.Key("master_port").SetValue("10888")
		shard.Key("cluster_key").SetValue("defaultkey")
	}

	misc := cfg.Section("MISC")
	if misc.Key("console_enabled").String() == "" {
		misc.Key("console_enabled").SetValue("true")
	}

	if err := cfg.SaveTo(filepath.Join(clusterDir, "cluster.ini")); err != nil {
		return err
	}

	if config.Token != "" {
		if err := os.WriteFile(filepath.Join(clusterDir, "cluster_token.txt"), []byte(config.Token), 0644); err != nil {
			return err
		}
	}

	return nil
}

// ReadShardPort reads server_port from a shard's server.ini
func ReadShardPort(shardDir string) int {
	cfg, err := ini.Load(filepath.Join(shardDir, "server.ini"))
	if err != nil {
		return 0
	}
	port, _ := cfg.Section("NETWORK").Key("server_port").Int()
	return port
}

// WriteShardPort writes server_port to a shard's server.ini
func WriteShardPort(shardDir string, port int) error {
	iniPath := filepath.Join(shardDir, "server.ini")
	cfg, err := ini.Load(iniPath)
	if err != nil {
		return err
	}
	cfg.Section("NETWORK").Key("server_port").SetValue(strconv.Itoa(port))
	return cfg.SaveTo(iniPath)
}

func InitClusterDir(clusterDir string, config *model.ClusterConfig, enableCaves bool) error {
	masterDir := filepath.Join(clusterDir, "Master")

	if err := os.MkdirAll(masterDir, 0755); err != nil {
		return err
	}

	// Write cluster.ini
	if err := os.WriteFile(filepath.Join(clusterDir, "cluster.ini"), []byte(DefaultClusterIni), 0644); err != nil {
		return err
	}
	if err := WriteClusterConfig(clusterDir, config); err != nil {
		return err
	}

	// Write Master shard config
	if err := os.WriteFile(filepath.Join(masterDir, "server.ini"), []byte(DefaultMasterServerIni), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(masterDir, "leveldataoverride.lua"), []byte(DefaultMasterLevelData), 0644); err != nil {
		return err
	}

	// Write Caves shard if enabled
	if enableCaves {
		cavesDir := filepath.Join(clusterDir, "Caves")
		if err := os.MkdirAll(cavesDir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(cavesDir, "server.ini"), []byte(DefaultCavesServerIni), 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(cavesDir, "leveldataoverride.lua"), []byte(DefaultCavesLevelData), 0644); err != nil {
			return err
		}
	}

	return nil
}
