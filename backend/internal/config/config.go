package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration structure loaded from config.yaml.
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	AI        AIConfig        `yaml:"ai"`
	Docker    DockerConfig    `yaml:"docker"`
	Clusters  []ClusterConfig `yaml:"clusters"`
	Registry  RegistryConfig  `yaml:"registry"`
	Database  DatabaseConfig  `yaml:"database"`
	Logging   LoggingConfig   `yaml:"logging"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	WebSocket WebSocketConfig `yaml:"websocket"`
	Security  SecurityConfig  `yaml:"security"`
	Features  FeaturesConfig  `yaml:"features"`
	Limits    LimitsConfig    `yaml:"limits"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type AIConfig struct {
	Provider    string       `yaml:"provider"`
	APIKey      string       `yaml:"api_key"`
	Model       string       `yaml:"model"`
	Temperature float64      `yaml:"temperature"`
	MaxTokens   int          `yaml:"max_tokens"`
	FewShot     FewShotConfig `yaml:"few_shot"`
	Cache       CacheConfig  `yaml:"cache"`
}

type FewShotConfig struct {
	Enabled             bool    `yaml:"enabled"`
	MaxExamples         int     `yaml:"max_examples"`
	SimilarityThreshold float64 `yaml:"similarity_threshold"`
}

type CacheConfig struct {
	Enabled bool          `yaml:"enabled"`
	TTL     time.Duration `yaml:"ttl"`
}

type DockerConfig struct {
	Local DockerLocalConfig `yaml:"local"`
}

type DockerLocalConfig struct {
	Socket string `yaml:"socket"`
}

type ClusterConfig struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"`
	Kubeconfig string `yaml:"kubeconfig"`
	Context    string `yaml:"context"`
	Registry   string `yaml:"registry"`
}

type RegistryConfig struct {
	Default RegistryCredentials `yaml:"default"`
}

type RegistryCredentials struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type DatabaseConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type MetricsConfig struct {
	Interval          int `yaml:"interval"`
	BroadcastInterval int `yaml:"broadcast_interval"`
	Retention         string `yaml:"retention"`
}

type WebSocketConfig struct {
	MaxConnections int           `yaml:"max_connections"`
	BufferSize     int           `yaml:"buffer_size"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	PingInterval   time.Duration `yaml:"ping_interval"`
}

type SecurityConfig struct {
	CORS CORSConfig `yaml:"cors"`
}

type CORSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

type FeaturesConfig struct {
	AIManifestGeneration bool `yaml:"ai_manifest_generation"`
	AutoDeploy           bool `yaml:"auto_deploy"`
	MetricsCollection    bool `yaml:"metrics_collection"`
	LogStreaming          bool `yaml:"log_streaming"`
	DeploymentHistory    bool `yaml:"deployment_history"`
}

type LimitsConfig struct {
	MaxConcurrentDeploys int `yaml:"max_concurrent_deploys"`
	DeployTimeout        int `yaml:"deploy_timeout"`
	MaxLogLines          int `yaml:"max_log_lines"`
}

// Load reads and parses the YAML configuration file at the given path.
// If path is empty, it falls back to the CONFIG_PATH environment variable.
func Load(path string) (*Config, error) {
	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}
	if path == "" {
		return nil, fmt.Errorf("config path is required: set CONFIG_PATH env or pass as argument")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Expand environment variables in the YAML content
	expanded := os.ExpandEnv(string(data))

	cfg := &Config{}
	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	applyDefaults(cfg)
	applyEnvOverrides(cfg)

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30 * time.Second
	}
	if cfg.Database.Type == "" {
		cfg.Database.Type = "sqlite"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./data/deployments.db"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Metrics.Interval == 0 {
		cfg.Metrics.Interval = 2
	}
	if cfg.Metrics.BroadcastInterval == 0 {
		cfg.Metrics.BroadcastInterval = 2
	}
	if cfg.WebSocket.MaxConnections == 0 {
		cfg.WebSocket.MaxConnections = 1000
	}
	if cfg.WebSocket.BufferSize == 0 {
		cfg.WebSocket.BufferSize = 256
	}
	if cfg.Limits.MaxConcurrentDeploys == 0 {
		cfg.Limits.MaxConcurrentDeploys = 5
	}
	if cfg.Limits.DeployTimeout == 0 {
		cfg.Limits.DeployTimeout = 30
	}
	if cfg.Limits.MaxLogLines == 0 {
		cfg.Limits.MaxLogLines = 10000
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" && cfg.AI.Provider == "openai" {
		cfg.AI.APIKey = v
	}
	if v := os.Getenv("CLAUDE_API_KEY"); v != "" && cfg.AI.Provider == "claude" {
		cfg.AI.APIKey = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("DATABASE_PATH"); v != "" {
		cfg.Database.Path = v
	}
	if v := os.Getenv("DOCKER_SOCKET"); v != "" {
		cfg.Docker.Local.Socket = v
	}
}
