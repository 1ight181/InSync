package models

import "errors"

// Реализует интерфейс ConfigModel
type DbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

var (
	ErrDbHostIsEmpty     = errors.New("Хост БД не может быть пустым")
	ErrDbPortIsInvalid   = errors.New("Порт БД должен быть целым числом в диапозоне от 0 до 65535")
	ErrDbUserIsEmpty     = errors.New("Пользователь БД не может быть пустым")
	ErrDbPasswordIsEmpty = errors.New("Пароль БД не может быть пустым")
)

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
