package errors

import "errors"

// Ошибки лоудера конфигурации
type FailedToLoadConfigError struct {
	Err error
}

func (e *FailedToLoadConfigError) Error() string {
	return "Не удалось загрузить конфигурацию: " + e.Err.Error()
}

// Ошибки валидации конфигурации

// Ошибки валидации конфигурации сервера
var (
	ErrServerIpIsEmpty          = errors.New("IP сервера не может быть пустым")
	ErrServerPortIsInvalid      = errors.New("Порт сервера должен быть целым числом в диапозоне от 0 до 65535")
	ErrServerNetworkTypeIsEmpty = errors.New("Тип сети сервера не может быть пустым")
)

// Ошибки валидации конфигурации базы данных
var (
	ErrDbHostIsEmpty     = errors.New("Хост БД не может быть пустым")
	ErrDbPortIsInvalid   = errors.New("Порт БД должен быть целым числом в диапозоне от 0 до 65535")
	ErrDbUserIsEmpty     = errors.New("Пользователь БД не может быть пустым")
	ErrDbPasswordIsEmpty = errors.New("Пароль БД не может быть пустым")
)

// Ошибки валидации конфигурации сертификатов
var (
	ErrServerCertDirIsEmpty       = errors.New("Директория сертификата сервера не может быть пустой")
	ErrServerCertFilenameIsEmpty  = errors.New("Имя файла сертификата сервера не может быть пустым")
	ErrServerCertExtensionIsEmpty = errors.New("Расширение файла сертификата сервера не может быть пустым")

	ErrServerKeyDirIsEmpty       = errors.New("Директория ключа сервера не может быть пустой")
	ErrServerKeyFilenameIsEmpty  = errors.New("Имя файла ключа сервера не может быть пустым")
	ErrServerKeyExtensionIsEmpty = errors.New("Расширение файла ключа сервера не может быть пустым")

	ErrClientCertDirIsEmpty       = errors.New("Директория сертификата клиента не может быть пустой")
	ErrClientCertFilenameIsEmpty  = errors.New("Имя файла сертификата клиента не может быть пустым")
	ErrClientCertExtensionIsEmpty = errors.New("Расширение файла сертификата клиента не может быть пустым")

	ErrClientKeyDirIsEmpty       = errors.New("Директория ключа клиента не может быть пустой")
	ErrClientKeyFilenameIsEmpty  = errors.New("Имя файла ключа клиента не может быть пустым")
	ErrClientKeyExtensionIsEmpty = errors.New("Расширение файла ключа клиента не может быть пустым")

	ErrCaCertDirIsEmpty       = errors.New("Директория сертификата CA не может быть пустой")
	ErrCaCertFilenameIsEmpty  = errors.New("Имя файла сертификата CA не может быть пустым")
	ErrCaCertExtensionIsEmpty = errors.New("Расширение файла сертификата CA не может быть пустым")
)

type ServerCertFileDoesNotExistError struct {
	FilePath string
}

func (e *ServerCertFileDoesNotExistError) Error() string {
	return "Файл сертификата сервера не существует: " + e.FilePath
}

type ServerKeyFileDoesNotExistError struct {
	FilePath string
}

func (e *ServerKeyFileDoesNotExistError) Error() string {
	return "Файл ключа сервера не существует: " + e.FilePath
}

type ClientCertFileDoesNotExistError struct {
	FilePath string
}

func (e *ClientCertFileDoesNotExistError) Error() string {
	return "Файл сертификата клиента не существует: " + e.FilePath
}

type ClientKeyFileDoesNotExistError struct {
	FilePath string
}

func (e *ClientKeyFileDoesNotExistError) Error() string {
	return "Файл ключа клиента не существует: " + e.FilePath
}

type CaCertFileDoesNotExistError struct {
	FilePath string
}

func (e *CaCertFileDoesNotExistError) Error() string {
	return "Файл сертификата CA не существует: " + e.FilePath
}
