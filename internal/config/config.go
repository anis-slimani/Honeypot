package config

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
)

// Config représente la configuration complète du honeypot
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Auth        AuthConfig        `yaml:"auth"`
	Shell       ShellConfig       `yaml:"shell"`
	Database    DatabaseConfig    `yaml:"database"`
	Logging     LoggingConfig     `yaml:"logging"`
	Web         WebConfig         `yaml:"web"`
	Alerts      AlertsConfig      `yaml:"alerts"`
	Geolocation GeolocationConfig `yaml:"geolocation"`
}

// ServerConfig configuration du serveur SSH
type ServerConfig struct {
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	MaxConnections   int    `yaml:"max_connections"`
	ConnectionTimeout int   `yaml:"connection_timeout"`
	IdleTimeout      int    `yaml:"idle_timeout"`
}

// AuthConfig configuration de l'authentification
type AuthConfig struct {
	FakeUsers []FakeUser `yaml:"fake_users"`
}

// FakeUser utilisateur factice
type FakeUser struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// ShellConfig configuration du shell factice
type ShellConfig struct {
	Prompt         string   `yaml:"prompt"`
	WelcomeMessage string   `yaml:"welcome_message"`
	FakeCommands   []string `yaml:"fake_commands"`
}

// DatabaseConfig configuration de la base de données
type DatabaseConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// LoggingConfig configuration du logging
type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

// WebConfig configuration de l'interface web
type WebConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	StaticPath    string `yaml:"static_path"`
	TemplatesPath string `yaml:"templates_path"`
}

// AlertsConfig configuration des alertes
type AlertsConfig struct {
	Enabled   bool        `yaml:"enabled"`
	Email     EmailConfig `yaml:"email"`
	Thresholds Thresholds `yaml:"thresholds"`
}

// EmailConfig configuration email
type EmailConfig struct {
	SMTPHost string   `yaml:"smtp_host"`
	SMTPPort int      `yaml:"smtp_port"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	From     string   `yaml:"from"`
	To       []string `yaml:"to"`
}

// Thresholds seuils d'alerte
type Thresholds struct {
	FailedAttempts int `yaml:"failed_attempts"`
	TimeWindow     int `yaml:"time_window"`
}

// GeolocationConfig configuration géolocalisation
type GeolocationConfig struct {
	Enabled bool   `yaml:"enabled"`
	APIKey  string `yaml:"api_key"`
}

// Load charge la configuration depuis un fichier YAML
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validation et valeurs par défaut
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validate valide la configuration et applique les valeurs par défaut
func (c *Config) validate() error {
	// Valeurs par défaut pour le serveur
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 2222
	}
	if c.Server.MaxConnections == 0 {
		c.Server.MaxConnections = 10
	}
	if c.Server.ConnectionTimeout == 0 {
		c.Server.ConnectionTimeout = 300
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 60
	}

	// Valeurs par défaut pour le shell
	if c.Shell.Prompt == "" {
		c.Shell.Prompt = "user@honeypot:~$ "
	}
	if c.Shell.WelcomeMessage == "" {
		c.Shell.WelcomeMessage = "Welcome to Ubuntu 20.04.3 LTS (GNU/Linux 5.4.0-74-generic x86_64)"
	}

	// Valeurs par défaut pour la base de données
	if c.Database.Type == "" {
		c.Database.Type = "sqlite"
	}
	if c.Database.Path == "" {
		c.Database.Path = "./honeypot.db"
	}

	// Valeurs par défaut pour le logging
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.File == "" {
		c.Logging.File = "./logs/honeypot.log"
	}

	// Valeurs par défaut pour le web
	if c.Web.Host == "" {
		c.Web.Host = "127.0.0.1"
	}
	if c.Web.Port == 0 {
		c.Web.Port = 8080
	}

	// Valeurs par défaut pour les alertes
	if c.Alerts.Thresholds.FailedAttempts == 0 {
		c.Alerts.Thresholds.FailedAttempts = 5
	}
	if c.Alerts.Thresholds.TimeWindow == 0 {
		c.Alerts.Thresholds.TimeWindow = 300
	}

	return nil
}
