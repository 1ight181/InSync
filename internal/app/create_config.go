package app

import (
	"context"
	cnf "insync/internal/infrastructure/config"
	cnfmodels "insync/internal/infrastructure/config/models"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	isDebug                     = true
	defaultConfigFileDirEnvKey  = "INSYNC_CONFIG_FILE_DIR"
	defaultConfigFileNameEnvKey = "INSYNC_CONFIG_FILE_NAME"
	defaultConfigFileDir        = "../config"
	defaultDebugConfigFileName  = "debug_config.yaml"
	defaultProdConfigFileName   = "config.yaml"
	defaultEnvPrefix            = "INSYNC"
	defaultEnvDelimiter         = "_"
	defaultKoanfDelimiter       = "."
	defaultTag                  = "mapstructure"
)

func getConfigFileInfoFromEnv() (string, string) {
	configFileDir := os.Getenv(defaultConfigFileDirEnvKey)
	configFileName := os.Getenv(defaultConfigFileNameEnvKey)

	return configFileDir, configFileName
}

func createConfig(ctx context.Context) (*cnfmodels.GeneralConfig, error) {
	configFileDir, configFileName := getConfigFileInfoFromEnv()
	if configFileDir == "" || configFileName == "" {
		if isDebug {
			configFileName = defaultDebugConfigFileName
		} else {
			configFileName = defaultProdConfigFileName
		}

		configFileDir = defaultConfigFileDir
	}

	stubLoggerHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	stubLogger := slog.New(stubLoggerHandler)

	configOpts := cnf.KoanfYamlEnvConfigLoaderOption{
		KoanfDelimiter: defaultKoanfDelimiter,
		YamlConfigFilePath: filepath.Join(
			configFileDir,
			configFileName,
		),
		EnvPrefix:    defaultEnvPrefix,
		EnvDelimiter: defaultEnvDelimiter,
		Logger:       stubLogger,
		Ctx:          ctx,
		Tag:          defaultTag,
	}

	configLoader := cnf.NewConfigLoader(configOpts)
	config, err := configLoader.LoadAndValidateConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}
