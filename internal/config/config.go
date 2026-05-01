package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Node  NodeConfig  `toml:"node"`
	Email EmailConfig `toml:"email"`
	P2P   P2PConfig   `toml:"p2p"`
}

type NodeConfig struct {
	Port    int    `toml:"port"`
	DataDir string `toml:"data_dir"`
	BaseURL string `toml:"base_url"`
	Public  bool   `toml:"public"`
	Dev     bool   `toml:"dev"`
	Verbose bool   `toml:"verbose"`
}

type EmailConfig struct {
	SMTPHost string `toml:"smtp_host"`
	SMTPPort int    `toml:"smtp_port"`
	SMTPUser string `toml:"smtp_user"`
	SMTPPass string `toml:"smtp_password"`
	From     string `toml:"from"`
}

type P2PConfig struct {
	BootstrapPeers []string `toml:"bootstrap_peers"`
	ListenPort     int      `toml:"listen_port"`
}

func Load(path string) (*Config, error) {
	cfg := defaults()
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err == nil {
		if _, err := toml.Decode(string(data), cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}
	return cfg, nil
}

func defaults() *Config {
	return &Config{
		Node: NodeConfig{
			Port:    7331,
			DataDir: "./data",
			Public:  true,
		},
		P2P: P2PConfig{
			ListenPort: 4001,
		},
	}
}
