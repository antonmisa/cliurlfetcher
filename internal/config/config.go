package config

import (
	"fmt"
	"os"
	"runtime"

	"github.com/ilyakaznacheev/cleanenv"
	"gopkg.in/yaml.v3"
)

// Config -.
type Config struct {
	Log `yaml:"logger"`
	App `yaml:"app"`
}

// App -.
type App struct {
	NumberOfWorkers int `env-required:"true" yaml:"workers"`
}

// Log -.
type Log struct {
	Level string `env-required:"true" yaml:"level" env:"LOG_LEVEL"`
	Path  string `env-required:"true" yaml:"path"`
}

func New() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yml"
	}

	cfg := &Config{}

	err := cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func Prepare() error {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yml"
	}

	if _, err := os.Stat(configPath); err == nil {
		return os.ErrExist
	}

	cfg := &Config{
		App: App{
			NumberOfWorkers: runtime.NumCPU(),
		},
		Log: Log{
			Level: "debug",
			Path:  "log.log",
		},
	}

	yamlData, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("error while marshaling config: %w", err)
	}

	err = os.WriteFile(configPath, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("error while creating config.yml: %w", err)
	}

	return nil
}
