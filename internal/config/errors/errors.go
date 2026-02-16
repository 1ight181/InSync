package errors

type FailedToLoadConfigError struct {
	Err error
}

func (e *FailedToLoadConfigError) Error() string {
	return "Не удалось загрузить конфигурацию: " + e.Err.Error()
}
