package models

// Реализует интерфейс ConfigModel
type GeneralConfig struct {
	DbConfig     DbConfig
	ServerConfig ServerConfig
	ClientConfig ClientConfig
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

	return nil
}
