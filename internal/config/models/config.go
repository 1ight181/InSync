package models

// Реализует интерфейс ConfigModel
type GeneralConfig struct {
	DbConfig     DbConfig
	ServerConfig ServerConfig
	ClientConfig ClientConfig
	LoggerConfig LoggerConfig
	TlsConfig    TlsConfig
}

func (gc *GeneralConfig) Validate() error {
	if err := gc.DbConfig.Validate(); err != nil {
		return err
	}
	if err := gc.ServerConfig.Validate(); err != nil {
		return err
	}
	if err := gc.ClientConfig.Validate(); err != nil {
		return err
	}
	if err := gc.LoggerConfig.Validate(); err != nil {
		return err
	}
	if err := gc.TlsConfig.Validate(); err != nil {
		return err
	}

	return nil
}
