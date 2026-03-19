package config

// Ошибки лоудера конфигурации
type FailedToLoadConfigError struct {
	Err error
}

func (e *FailedToLoadConfigError) Error() string {
	return "Не удалось загрузить конфигурацию: " + e.Err.Error()
}
