// Package config provides configuration management for Krustron
// Author: Anubhav Gain <anubhavg@infopercept.com>
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	NATS        NATSConfig        `mapstructure:"nats"`
	Auth        AuthConfig        `mapstructure:"auth"`
	Kubernetes  KubernetesConfig  `mapstructure:"kubernetes"`
	GitOps      GitOpsConfig      `mapstructure:"gitops"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Security    SecurityConfig    `mapstructure:"security"`
	AI          AIConfig          `mapstructure:"ai"`
	Logger      LoggerConfig      `mapstructure:"logger"`
}

// ServerConfig holds HTTP/gRPC server configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	GRPCPort        int           `mapstructure:"grpc_port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	Mode            string        `mapstructure:"mode"` // debug, release, test
	CorsOrigins     []string      `mapstructure:"cors_origins"`
	TLSEnabled      bool          `mapstructure:"tls_enabled"`
	TLSCert         string        `mapstructure:"tls_cert"`
	TLSKey          string        `mapstructure:"tls_key"`
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	MigrationsPath  string        `mapstructure:"migrations_path"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Password    string        `mapstructure:"password"`
	DB          int           `mapstructure:"db"`
	PoolSize    int           `mapstructure:"pool_size"`
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
}

// NATSConfig holds NATS messaging configuration
type NATSConfig struct {
	URL            string        `mapstructure:"url"`
	ClusterID      string        `mapstructure:"cluster_id"`
	ClientID       string        `mapstructure:"client_id"`
	MaxReconnects  int           `mapstructure:"max_reconnects"`
	ReconnectWait  time.Duration `mapstructure:"reconnect_wait"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret           string        `mapstructure:"jwt_secret"`
	JWTExpiration       time.Duration `mapstructure:"jwt_expiration"`
	RefreshExpiration   time.Duration `mapstructure:"refresh_expiration"`
	OIDCEnabled         bool          `mapstructure:"oidc_enabled"`
	OIDCIssuer          string        `mapstructure:"oidc_issuer"`
	OIDCClientID        string        `mapstructure:"oidc_client_id"`
	OIDCClientSecret    string        `mapstructure:"oidc_client_secret"`
	OIDCRedirectURL     string        `mapstructure:"oidc_redirect_url"`
	CasbinModelPath     string        `mapstructure:"casbin_model_path"`
	CasbinPolicyPath    string        `mapstructure:"casbin_policy_path"`
	SessionSecret       string        `mapstructure:"session_secret"`
	BCryptCost          int           `mapstructure:"bcrypt_cost"`
}

// KubernetesConfig holds Kubernetes client configuration
type KubernetesConfig struct {
	InCluster           bool          `mapstructure:"in_cluster"`
	KubeconfigPath      string        `mapstructure:"kubeconfig_path"`
	DefaultNamespace    string        `mapstructure:"default_namespace"`
	WatchResyncPeriod   time.Duration `mapstructure:"watch_resync_period"`
	QPS                 float32       `mapstructure:"qps"`
	Burst               int           `mapstructure:"burst"`
	AgentImage          string        `mapstructure:"agent_image"`
	AgentNamespace      string        `mapstructure:"agent_namespace"`
}

// GitOpsConfig holds GitOps configuration
type GitOpsConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	Provider          string        `mapstructure:"provider"` // argocd, flux
	ArgoCD            ArgoCDConfig  `mapstructure:"argocd"`
	SyncInterval      time.Duration `mapstructure:"sync_interval"`
	PruneEnabled      bool          `mapstructure:"prune_enabled"`
	SelfHealEnabled   bool          `mapstructure:"self_heal_enabled"`
}

// ArgoCDConfig holds ArgoCD specific configuration
type ArgoCDConfig struct {
	ServerURL   string `mapstructure:"server_url"`
	AuthToken   string `mapstructure:"auth_token"`
	Insecure    bool   `mapstructure:"insecure"`
	Namespace   string `mapstructure:"namespace"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	Metrics     MetricsConfig     `mapstructure:"metrics"`
	Tracing     TracingConfig     `mapstructure:"tracing"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Prometheus  PrometheusConfig  `mapstructure:"prometheus"`
	Grafana     GrafanaConfig     `mapstructure:"grafana"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Path      string `mapstructure:"path"`
	Namespace string `mapstructure:"namespace"`
}

// TracingConfig holds distributed tracing configuration
type TracingConfig struct {
	Enabled     bool    `mapstructure:"enabled"`
	Provider    string  `mapstructure:"provider"` // jaeger, otel
	Endpoint    string  `mapstructure:"endpoint"`
	ServiceName string  `mapstructure:"service_name"`
	SampleRate  float64 `mapstructure:"sample_rate"`
}

// LoggingConfig holds centralized logging configuration
type LoggingConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Provider  string `mapstructure:"provider"` // opensearch, elasticsearch
	Endpoint  string `mapstructure:"endpoint"`
	Index     string `mapstructure:"index"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
}

// PrometheusConfig holds Prometheus configuration
type PrometheusConfig struct {
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// GrafanaConfig holds Grafana configuration
type GrafanaConfig struct {
	URL      string `mapstructure:"url"`
	APIKey   string `mapstructure:"api_key"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	TrivyEnabled    bool          `mapstructure:"trivy_enabled"`
	TrivyServerURL  string        `mapstructure:"trivy_server_url"`
	OPAEnabled      bool          `mapstructure:"opa_enabled"`
	OPAServerURL    string        `mapstructure:"opa_server_url"`
	WazuhEnabled    bool          `mapstructure:"wazuh_enabled"`
	WazuhServerURL  string        `mapstructure:"wazuh_server_url"`
	WazuhAPIKey     string        `mapstructure:"wazuh_api_key"`
	ScanInterval    time.Duration `mapstructure:"scan_interval"`
	BlockOnCritical bool          `mapstructure:"block_on_critical"`
}

// AIConfig holds AI/LLM configuration
type AIConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	Provider     string `mapstructure:"provider"` // ollama, openai, anthropic
	Endpoint     string `mapstructure:"endpoint"`
	Model        string `mapstructure:"model"`
	APIKey       string `mapstructure:"api_key"`
	MaxTokens    int    `mapstructure:"max_tokens"`
	Temperature  float64 `mapstructure:"temperature"`
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Format      string `mapstructure:"format"`
	Output      string `mapstructure:"output"`
	Development bool   `mapstructure:"development"`
}

// Load loads configuration from file and environment
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/krustron")
		v.AddConfigPath("$HOME/.krustron")
	}

	// Environment variables
	v.SetEnvPrefix("KRUSTRON")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables for sensitive data
	overrideFromEnv(&cfg)

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.grpc_port", 50051)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.shutdown_timeout", "10s")
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.cors_origins", []string{"*"})

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "krustron")
	v.SetDefault("database.database", "krustron")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "5m")
	v.SetDefault("database.migrations_path", "migrations")

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.dial_timeout", "5s")

	// NATS defaults
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.cluster_id", "krustron-cluster")
	v.SetDefault("nats.client_id", "krustron-api")
	v.SetDefault("nats.max_reconnects", 10)
	v.SetDefault("nats.reconnect_wait", "2s")
	v.SetDefault("nats.connect_timeout", "10s")

	// Auth defaults
	v.SetDefault("auth.jwt_expiration", "24h")
	v.SetDefault("auth.refresh_expiration", "168h")
	v.SetDefault("auth.bcrypt_cost", 12)
	v.SetDefault("auth.casbin_model_path", "configs/casbin_model.conf")
	v.SetDefault("auth.casbin_policy_path", "configs/casbin_policy.csv")

	// Kubernetes defaults
	v.SetDefault("kubernetes.in_cluster", false)
	v.SetDefault("kubernetes.default_namespace", "default")
	v.SetDefault("kubernetes.watch_resync_period", "30m")
	v.SetDefault("kubernetes.qps", 50)
	v.SetDefault("kubernetes.burst", 100)
	v.SetDefault("kubernetes.agent_image", "ghcr.io/anubhavg-icpl/krustron-agent:latest")
	v.SetDefault("kubernetes.agent_namespace", "krustron-system")

	// GitOps defaults
	v.SetDefault("gitops.enabled", true)
	v.SetDefault("gitops.provider", "argocd")
	v.SetDefault("gitops.sync_interval", "3m")
	v.SetDefault("gitops.prune_enabled", true)
	v.SetDefault("gitops.self_heal_enabled", true)

	// Observability defaults
	v.SetDefault("observability.metrics.enabled", true)
	v.SetDefault("observability.metrics.path", "/metrics")
	v.SetDefault("observability.metrics.namespace", "krustron")
	v.SetDefault("observability.tracing.enabled", false)
	v.SetDefault("observability.tracing.provider", "otel")
	v.SetDefault("observability.tracing.sample_rate", 0.1)

	// Security defaults
	v.SetDefault("security.trivy_enabled", true)
	v.SetDefault("security.opa_enabled", true)
	v.SetDefault("security.scan_interval", "1h")
	v.SetDefault("security.block_on_critical", true)

	// AI defaults
	v.SetDefault("ai.enabled", false)
	v.SetDefault("ai.provider", "ollama")
	v.SetDefault("ai.endpoint", "http://localhost:11434")
	v.SetDefault("ai.model", "llama3")
	v.SetDefault("ai.max_tokens", 2048)
	v.SetDefault("ai.temperature", 0.7)

	// Logger defaults
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.output", "stdout")
	v.SetDefault("logger.development", false)
}

// overrideFromEnv overrides sensitive config from environment variables
func overrideFromEnv(cfg *Config) {
	if v := os.Getenv("KRUSTRON_DATABASE_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("KRUSTRON_REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("KRUSTRON_AUTH_JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}
	if v := os.Getenv("KRUSTRON_AUTH_SESSION_SECRET"); v != "" {
		cfg.Auth.SessionSecret = v
	}
	if v := os.Getenv("KRUSTRON_AUTH_OIDC_CLIENT_SECRET"); v != "" {
		cfg.Auth.OIDCClientSecret = v
	}
	if v := os.Getenv("KRUSTRON_GITOPS_ARGOCD_AUTH_TOKEN"); v != "" {
		cfg.GitOps.ArgoCD.AuthToken = v
	}
	if v := os.Getenv("KRUSTRON_SECURITY_WAZUH_API_KEY"); v != "" {
		cfg.Security.WazuhAPIKey = v
	}
	if v := os.Getenv("KRUSTRON_AI_API_KEY"); v != "" {
		cfg.AI.APIKey = v
	}
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// RedisAddr returns the Redis address
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ServerAddr returns the server address
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GRPCAddr returns the gRPC server address
func (c *ServerConfig) GRPCAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.GRPCPort)
}
