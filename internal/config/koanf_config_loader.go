package config

import (
	conferr "insync/internal/config/errors"
	confifaces "insync/internal/config/interfaces"
	confmodels "insync/internal/config/models"
	"strings"

	slog "log/slog"

	yamlparser "github.com/knadh/koanf/parsers/yaml"
	envprovider "github.com/knadh/koanf/providers/env"
	fileprovider "github.com/knadh/koanf/providers/file"
	koanf "github.com/knadh/koanf/v2"
)

type KoanfYamlEnvConfigLoader struct {
	koanfDelimiter     string
	yamlConfigFilePath string
	envPrefix          string
	envDelimiter       string
	logger             *slog.Logger
}

type KoanfYamlEnvConfigLoaderOption struct {
	KoanfDelimiter     string
	YamlConfigFilePath string
	Logger             *slog.Logger
	EnvPrefix          string
	EnvDelimiter       string
}

func NewConfigLoader(opts KoanfYamlEnvConfigLoaderOption) confifaces.IConfigLoader[*confmodels.GeneralConfig] {
	if opts.KoanfDelimiter == "" ||
		opts.YamlConfigFilePath == "" ||
		opts.Logger == nil ||
		opts.EnvPrefix == "" ||
		opts.EnvDelimiter == "" {
		panic("Все поля KoanfYamlEnvConfigLoaderOption должны быть заполнены")
	}
	return &KoanfYamlEnvConfigLoader{
		koanfDelimiter:     opts.KoanfDelimiter,
		yamlConfigFilePath: opts.YamlConfigFilePath,
		logger:             opts.Logger,
		envPrefix:          opts.EnvPrefix,
		envDelimiter:       opts.EnvDelimiter,
	}
}

func (kyecl *KoanfYamlEnvConfigLoader) LoadAndValidateConfig() (*confmodels.GeneralConfig, error) {
	koanfLoader := koanf.New(kyecl.koanfDelimiter)

	fileProvider := fileprovider.Provider(kyecl.yamlConfigFilePath)

	yamlParser := yamlparser.Parser()

	if err := koanfLoader.Load(fileProvider, yamlParser); err != nil {
		return nil, &conferr.FailedToLoadConfigError{Err: err}
	}

	envProvider := envprovider.Provider(
		kyecl.envPrefix,
		kyecl.envDelimiter,
		func(envVarName string) string {
			envVarNameWithKoanfDelimiter := strings.ReplaceAll(envVarName, kyecl.envDelimiter, kyecl.koanfDelimiter)
			lowerEnvVarName := strings.ToLower(envVarNameWithKoanfDelimiter)

			return lowerEnvVarName
		},
	)

	if err := koanfLoader.Load(envProvider, nil); err != nil {
		return nil, &conferr.FailedToLoadConfigError{Err: err}
	}

	var generalConfig confmodels.GeneralConfig
	if err := koanfLoader.Unmarshal("", &generalConfig); err != nil {
		return nil, &conferr.FailedToLoadConfigError{Err: err}
	}

	return &generalConfig, nil
}
