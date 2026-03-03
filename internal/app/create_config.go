package app

import (
	cnf "insync/internal/config"
	cnfmodels "insync/internal/config/models"
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
)

func getConfigFileInfoFromEnv() (string, string) {
	configFileDir := os.Getenv(defaultConfigFileDirEnvKey)
	configFileName := os.Getenv(defaultConfigFileNameEnvKey)

	return configFileDir, configFileName
}

func createConfig() (*cnfmodels.GeneralConfig, error) {
	configFileDir, configFileName := getConfigFileInfoFromEnv()
	if configFileDir == "" || configFileName == "" {
		if isDebug {
			configFileName = defaultDebugConfigFileName
		} else {
			configFileName = defaultProdConfigFileName
		}

		configFileDir = defaultConfigFileDir
	}

	configOpts := cnf.KoanfYamlEnvConfigLoaderOption{
		KoanfDelimiter: ".",
		YamlConfigFilePath: filepath.Join(
			configFileDir,
			configFileName,
		),
		Logger:       slog.Default(),
		EnvPrefix:    defaultEnvPrefix,
		EnvDelimiter: defaultEnvDelimiter,
	}

	configLoader := cnf.NewConfigLoader(configOpts)
	config, err := configLoader.LoadAndValidateConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}
