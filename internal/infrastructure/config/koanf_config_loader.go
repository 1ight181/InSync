package config

import (
	"errors"
	"insync/internal/infrastructure/config/models"
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
	tag                string
}

type KoanfYamlEnvConfigLoaderOption struct {
	KoanfDelimiter     string
	YamlConfigFilePath string
	EnvPrefix          string
	EnvDelimiter       string
	Logger             *slog.Logger
	Tag                string
}

var (
	ErrInvalidOpts = errors.New("Все поля должны быть заполнены")
)

func NewConfigLoader(opts KoanfYamlEnvConfigLoaderOption) (*KoanfYamlEnvConfigLoader, error) {
	if opts.KoanfDelimiter == "" ||
		opts.YamlConfigFilePath == "" ||
		opts.EnvPrefix == "" ||
		opts.EnvDelimiter == "" ||
		opts.Logger == nil ||
		opts.Tag == "" {
		return nil, ErrInvalidOpts
	}
	return &KoanfYamlEnvConfigLoader{
		koanfDelimiter:     opts.KoanfDelimiter,
		yamlConfigFilePath: opts.YamlConfigFilePath,
		logger:             opts.Logger,
		envPrefix:          opts.EnvPrefix,
		envDelimiter:       opts.EnvDelimiter,
		tag:                opts.Tag,
	}, nil
}

func (kyecl *KoanfYamlEnvConfigLoader) LoadAndValidateConfig() (*models.GeneralConfig, error) {
	koanfLoader := koanf.New(kyecl.koanfDelimiter)

	fileProvider := fileprovider.Provider(kyecl.yamlConfigFilePath)

	yamlParser := yamlparser.Parser()

	if err := koanfLoader.Load(fileProvider, yamlParser); err != nil {
		return nil, &FailedToLoadConfigError{Err: err}
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
		return nil, &FailedToLoadConfigError{Err: err}
	}

	var generalConfig models.GeneralConfig
	if err := koanfLoader.UnmarshalWithConf("", &generalConfig, koanf.UnmarshalConf{
		Tag: kyecl.tag,
	}); err != nil {
		return nil, &FailedToLoadConfigError{Err: err}
	}

	if err := generalConfig.Validate(); err != nil {
		return nil, &FailedToLoadConfigError{Err: err}
	}

	kyecl.logger.Info("Конфигурация успешно загружена")

	return &generalConfig, nil
}
