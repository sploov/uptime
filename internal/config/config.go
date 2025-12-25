package config

import (
	"os"
	"time"

	"github.com/sploov/uptime/internal/models"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Targets []models.ServiceConfig `yaml:"targets"`
	Discord DiscordConfig          `yaml:"discord"`
}

type DiscordConfig struct {
	Enabled     bool   `yaml:"enabled"`
	WebhookURL  string `yaml:"webhook_url"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	// Set defaults if needed
	for i := range cfg.Targets {
		if cfg.Targets[i].Interval == 0 {
			cfg.Targets[i].Interval = 30 * time.Second // Default interval
		}
		if cfg.Targets[i].Timeout == 0 {
			cfg.Targets[i].Timeout = 5 * time.Second
		}
		if cfg.Targets[i].Method == "" {
			cfg.Targets[i].Method = "HTTP"
		}
	}

	return &cfg, nil
}

