package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"     env:",prefix=APP_"`
	Database   DatabaseConfig   `yaml:"database"   env:",prefix=DB_"`
	Logging    LoggingConfig    `yaml:"logging"    env:",prefix=LOG_"`
	Generation GenerationConfig `yaml:"generation"`
}

type ServerConfig struct {
	Port int `yaml:"port" env:"PORT" envDefault:"8080"`
}

type DatabaseConfig struct {
	Driver   string `yaml:"driver"   env:"DRIVER"   envDefault:"postgres"`
	Host     string `yaml:"host"     env:"HOST"     envDefault:"localhost"`
	Port     int    `yaml:"port"     env:"PORT"     envDefault:"5432"`
	User     string `yaml:"user"     env:"USER"     envDefault:"postgres"`
	Password string `yaml:"password" env:"PASSWORD" envDefault:""`
	DBName   string `yaml:"dbname"   env:"NAME"     envDefault:"app"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"  env:"LEVEL"  envDefault:"info"`
	Format string `yaml:"format" env:"FORMAT" envDefault:"json"`
}

type GenerationConfig struct {
	ORM       string `yaml:"orm"       envDefault:"gorm"`
	Testing   bool   `yaml:"testing"   envDefault:"true"`
	Benchmark bool   `yaml:"benchmark" envDefault:"true"`
}

func (d DatabaseConfig) DSN() string {
	switch d.Driver {
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			d.Host, d.Port, d.User, d.Password, d.DBName,
		)
	case "mysql", "mariadb":
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=True&loc=Local",
			d.User, d.Password, d.Host, d.Port, d.DBName,
		)
	case "sqlserver":
		return fmt.Sprintf(
			"sqlserver://%s:%s@%s:%d?database=%s",
			d.User, d.Password, d.Host, d.Port, d.DBName,
		)
	case "sqlite":
		return d.DBName
	default:
		return ""
	}
}

// Load applies config in priority order: YAML < ENV.
// CLI overrides are applied by the caller after Load returns.
func Load(path string) (*Config, error) {
	cfg := defaultConfig()

	if path != "" {
		if err := loadYAML(path, cfg); err != nil {
			return nil, fmt.Errorf("load yaml config: %w", err)
		}
	}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env config: %w", err)
	}

	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Server:   ServerConfig{Port: 8080},
		Database: DatabaseConfig{Driver: "postgres", Host: "localhost", Port: 5432},
		Logging:  LoggingConfig{Level: "info", Format: "json"},
		Generation: GenerationConfig{
			ORM:       "gorm",
			Testing:   true,
			Benchmark: true,
		},
	}
}

func loadYAML(path string, dst *Config) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(dst)
}
