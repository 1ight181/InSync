package models

// Реализует интерфейс ConfigModel
type DbConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

func (dc *DbConfig) Validate() error {
	if dc.Host == "" {
		return ErrDbHostIsEmpty
	}
	if dc.Port < 0 || dc.Port > 65535 {
		return ErrDbPortIsInvalid
	}
	if dc.User == "" {
		return ErrDbUserIsEmpty
	}
	if dc.Password == "" {
		return ErrDbPasswordIsEmpty
	}

	return nil
}
