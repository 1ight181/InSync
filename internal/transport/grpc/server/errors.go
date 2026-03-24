package server

import (
	"errors"
)

var (
	ErrFailedToAppendCa     = errors.New("не удалось добавить CA сертификат в пул")
	ErrServerAlreadyStarted = errors.New("сервер уже запущен")
	ErrServerAlreadyStopped = errors.New("сервер уже остановлен")
)
