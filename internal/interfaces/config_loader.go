package interfaces

type IConfigLoader[configModel IConfigModel] interface {
	LoadAndValidateConfig() (configModel, error)
}
