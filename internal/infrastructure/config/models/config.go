package models

// Реализует интерфейс ConfigModel
type GeneralConfig struct {
	DbConfig                     DbConfig                     `mapstructure:"db"`
	ServerConfig                 ServerConfig                 `mapstructure:"server"`
	ClientConfig                 ClientConfig                 `mapstructure:"client"`
	LoggerConfig                 LoggerConfig                 `mapstructure:"logger"`
	MDnsServerConfig             MDnsServerConfig             `mapstructure:"mdns_server"`
	MDnsBrowserConfig            MDnsBrowserConfig            `mapstructure:"mdns_browser"`
	FileManagerConfig            FileManagerConfig            `mapstructure:"file_manager"`
	RemoteDeviceIdResolverConfig RemoteDeviceIdResolverConfig `mapstructure:"remote_device_id_resolver"`
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
	if err := gc.MDnsServerConfig.Validate(); err != nil {
		return err
	}
	if err := gc.MDnsBrowserConfig.Validate(); err != nil {
		return err
	}
	if err := gc.FileManagerConfig.Validate(); err != nil {
		return err
	}
	if err := gc.RemoteDeviceIdResolverConfig.Validate(); err != nil {
		return err
	}

	return nil
}
