package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	defaultFile         = ".env"
	defaultOverrideFile = ".local.env"
)

type EnvLoader struct {
	logger logger
}

func New(configPath string, logger logger) Config {
	cfg := &EnvLoader{logger: logger}
	cfg.load(configPath)
	return cfg
}

// Load loads the environment variables from the given configPath
func (e *EnvLoader) load(configPath string) {
	var (
		defaultFile  = configPath + defaultFile
		overrideFile = configPath + defaultOverrideFile
		env          = e.GetString("APP_ENV", "")
	)

	err := godotenv.Load(defaultFile)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			e.logger.Fatalf("Error loading %s: %v", defaultFile, err)
		}
		e.logger.Warnf("No %s file found", defaultFile)
	} else {
		e.logger.Infof("Loaded %s", defaultFile)
	}

	if env != "" {
		overrideFile = fmt.Sprintf("%s/.%s.env", configPath, env)
	}

	err = godotenv.Overload(overrideFile)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			e.logger.Fatalf("Error loading %s: %v", overrideFile, err)
		}
		e.logger.Warnf("No %s file found", overrideFile)
	} else {
		e.logger.Infof("Loaded %s", overrideFile)
	}
}

// GetString returns the env variable for the given key
// and falls back to the given defaultValue if not set
func (e *EnvLoader) GetString(key, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if ok {
		return v
	}
	return defaultValue
}

// GetInt returns the env variable (parsed as integer) for
// the given key and falls back to the given defaultValue if not set
func (e *EnvLoader) GetInt(key string, defaultValue int) (int, error) {
	v, ok := os.LookupEnv(key)
	if ok {
		value, err := strconv.Atoi(v)
		if err != nil {
			return defaultValue, err
		}
		return value, nil
	}
	return defaultValue, nil
}

// GetFloat64 returns the env variable (parsed as float64) for
// the given key and falls back to the given defaultValue if not set
func (e *EnvLoader) GetFloat64(key string, defaultValue float64) (float64, error) {
	v, ok := os.LookupEnv(key)
	if ok {
		value, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return defaultValue, err
		}
		return value, nil
	}
	return defaultValue, nil
}

// GetBool returns the env variable (parsed as bool) for
// the given key and falls back to the given defaultValue if not set
func (e *EnvLoader) GetBool(key string, defaultValue bool) (bool, error) {
	v, ok := os.LookupEnv(key)
	if ok {
		value, err := strconv.ParseBool(v)
		if err != nil {
			return defaultValue, err
		}
		return value, nil
	}
	return defaultValue, nil
}
