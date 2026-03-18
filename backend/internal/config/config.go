package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Env         string `yaml:"env"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	AutoMigrate bool   `yaml:"auto_migrate"`
}

type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	Params   string `yaml:"params"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func Load(path string) (Config, error) {
	cfg := defaultConfig()

	if path != "" {
		if content, err := os.ReadFile(path); err == nil {
			if err := yaml.Unmarshal(content, &cfg); err != nil {
				return Config{}, fmt.Errorf("unmarshal config yaml: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("read config file: %w", err)
		}
	}

	applyEnvOverrides(&cfg)

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		App: AppConfig{
			Name:        "memory-system-backend",
			Env:         "development",
			Host:        "0.0.0.0",
			Port:        8080,
			AutoMigrate: true,
		},
		Database: DatabaseConfig{
			Driver:   "mysql",
			Host:     "127.0.0.1",
			Port:     3306,
			User:     "root",
			Password: "memory_password",
			Name:     "memory_system",
			Params:   "charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		},
		Redis: RedisConfig{
			Host: "127.0.0.1",
			Port: 6379,
			DB:   0,
		},
	}
}

func applyEnvOverrides(cfg *Config) {
	cfg.App.Name = getEnv("APP_NAME", cfg.App.Name)
	cfg.App.Env = getEnv("APP_ENV", cfg.App.Env)
	cfg.App.Host = getEnv("SERVER_HOST", cfg.App.Host)
	cfg.App.Port = getEnvAsInt("SERVER_PORT", cfg.App.Port)
	cfg.App.AutoMigrate = getEnvAsBool("AUTO_MIGRATE", cfg.App.AutoMigrate)

	cfg.Database.Driver = getEnv("MYSQL_DRIVER", cfg.Database.Driver)
	cfg.Database.Host = getEnv("MYSQL_HOST", cfg.Database.Host)
	cfg.Database.Port = getEnvAsInt("MYSQL_PORT", cfg.Database.Port)
	cfg.Database.User = getEnv("MYSQL_USER", cfg.Database.User)
	cfg.Database.Password = getEnv("MYSQL_PASSWORD", cfg.Database.Password)
	cfg.Database.Name = getEnv("MYSQL_DATABASE", cfg.Database.Name)
	cfg.Database.Params = getEnv("MYSQL_PARAMS", cfg.Database.Params)

	cfg.Redis.Host = getEnv("REDIS_HOST", cfg.Redis.Host)
	cfg.Redis.Port = getEnvAsInt("REDIS_PORT", cfg.Redis.Port)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", cfg.Redis.Password)
	cfg.Redis.DB = getEnvAsInt("REDIS_DB", cfg.Redis.DB)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func (c AppConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", c.User, c.Password, c.Host, c.Port, c.Name, c.Params)
}

func (c DatabaseConfig) MigrationURL() string {
	return fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s?%s", c.User, c.Password, c.Host, c.Port, c.Name, c.Params)
}

func (c RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

