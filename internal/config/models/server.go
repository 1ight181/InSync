package models

import "errors"

// Реализует интерфейс ConfigModel
type ServerConfig struct {
	Ip          string
	Port        int
	NetworkType string
}

var (
	ErrServerIpIsEmpty          = errors.New("IP сервера не может быть пустым")
	ErrServerPortIsInvalid      = errors.New("Порт сервера должен быть целым числом в диапозоне от 0 до 65535")
	ErrServerNetworkTypeIsEmpty = errors.New("Тип сети сервера не может быть пустым")
)

func (sc *ServerConfig) Validate() error {
	if sc.Ip == "" {
		return ErrServerIpIsEmpty
	}
	if sc.Port < 0 || sc.Port > 65535 {
		return ErrServerPortIsInvalid
	}
	if sc.NetworkType == "" {
		return ErrServerNetworkTypeIsEmpty
	}

	return nil
}
