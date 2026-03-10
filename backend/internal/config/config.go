package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port           string `json:"port"`
	DataDir        string `json:"dataDir"`
	ImageName      string `json:"imageName"`
	Platform       string `json:"platform"`
	Auth           Auth   `json:"auth"`
	BackupInterval int    `json:"backupInterval"` // Auto-backup interval in hours, 0 = disabled
	DiscordWebhook string `json:"discordWebhook"` // Discord webhook URL for notifications
}

type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Secret   string `json:"secret"` // JWT signing secret
}

func DefaultConfig() *Config {
	return &Config{
		Port:      "8080",
		DataDir:   "./data",
		ImageName: "dst-server:latest",
		Platform:  "linux/amd64",
		Auth: Auth{
			Username: "admin",
			Password: "admin",
			Secret:   "dst-ds-panel-secret-change-me",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	cfg.applyEnv()
	return cfg, nil
}

func (c *Config) applyEnv() {
	if v := os.Getenv("PORT"); v != "" {
		c.Port = v
	}
	if v := os.Getenv("DATA_DIR"); v != "" {
		c.DataDir = v
	}
	if v := os.Getenv("DST_IMAGE"); v != "" {
		c.ImageName = v
	}
	if v := os.Getenv("DST_PLATFORM"); v != "" {
		c.Platform = v
	}
	if v := os.Getenv("AUTH_USERNAME"); v != "" {
		c.Auth.Username = v
	}
	if v := os.Getenv("AUTH_PASSWORD"); v != "" {
		c.Auth.Password = v
	}
	if v := os.Getenv("AUTH_SECRET"); v != "" {
		c.Auth.Secret = v
	}
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
