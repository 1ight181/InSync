package config

import (
	"context"
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
	ctx                context.Context
	tag                string
}

type KoanfYamlEnvConfigLoaderOption struct {
	KoanfDelimiter     string
	YamlConfigFilePath string
	EnvPrefix          string
	EnvDelimiter       string
	Logger             *slog.Logger
	Ctx                context.Context
	Tag                string
}

func NewConfigLoader(opts KoanfYamlEnvConfigLoaderOption) confifaces.IConfigLoader[*confmodels.GeneralConfig] {
	if opts.KoanfDelimiter == "" ||
		opts.YamlConfigFilePath == "" ||
		opts.EnvPrefix == "" ||
		opts.EnvDelimiter == "" ||
		opts.Logger == nil ||
		opts.Ctx == nil ||
		opts.Tag == "" {
		panic("Все поля KoanfYamlEnvConfigLoaderOption должны быть заполнены")
	}
	return &KoanfYamlEnvConfigLoader{
		koanfDelimiter:     opts.KoanfDelimiter,
		yamlConfigFilePath: opts.YamlConfigFilePath,
		logger:             opts.Logger,
		envPrefix:          opts.EnvPrefix,
		envDelimiter:       opts.EnvDelimiter,
		ctx:                opts.Ctx,
		tag:                opts.Tag,
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
	if err := koanfLoader.UnmarshalWithConf("", &generalConfig, koanf.UnmarshalConf{
		Tag: kyecl.tag,
	}); err != nil {
		return nil, &conferr.FailedToLoadConfigError{Err: err}
	}

	if err := generalConfig.Validate(); err != nil {
		return nil, &conferr.FailedToLoadConfigError{Err: err}
	}

	kyecl.logger.Info("Конфигурация успешно загружена")

	return &generalConfig, nil
}
