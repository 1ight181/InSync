package config

type IConfigLoader[configModel IConfigModel] interface {
	LoadAndValidateConfig() (configModel, error)
}
