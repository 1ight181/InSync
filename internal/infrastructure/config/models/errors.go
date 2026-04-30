package models

import "errors"

// Ошибки валидации конфигурации сервера
var (
	ErrServerPortIsInvalid      = errors.New("Порт сервера должен быть целым числом в диапозоне от 0 до 65535")
	ErrServerNetworkTypeIsEmpty = errors.New("Тип сети сервера не может быть пустым")
)

// Ошибки валидации конфигурации хеша
var (
	ErrInvalidHashCacheEntryExpireUnixTime = errors.New("Время жизни хеша в секундах должно быть положительным целым числом")
)

// Ошибки валидации конфигурации клиента
var (
	ErrInvalidOptsForMdnsScheme    = errors.New("Неверные опции для схемы mDNS")
	ErrAddressToConnectIsEmpty     = errors.New("Адрес сервера, к которому нужно подключиться, не может быть пустым")
	ErrPortToConnectIsInvalid      = errors.New("Порт сервера, к которому нужно подключиться, должен быть целым числом в диапозоне от 0 до 65535")
	ErrNetworkToConnectTypeIsEmpty = errors.New("Тип сети сервера, к которому нужно подключиться, не может быть пустым")
	ErrClientResolverSchemeIsEmpty = errors.New("Схема резолвера не может быть пустой")
	ErrChunkSizeIsInvalid          = errors.New("Размер чанка должен быть положительным целым числом")
	ErrLoadBalancingPolicyIsEmpty  = errors.New("Политика балансировки не может быть пустой")
	ErrRpcTimeoutIsInvalid         = errors.New("Таймаут RPC должен быть положительным целым числом")
)

// Ошибки валидации конфигурации RpcRetryPolicy
var (
	ErrInvalidMaxAttempts           = errors.New("Максимальное количество попыток должно быть положительным целым числом")
	ErrInvalidInitialBackoffSeconds = errors.New("Начальное время ожидания должно быть положительным целым числом")
	ErrInvalidMaxBackoffSeconds     = errors.New("Максимальное время ожидания должно быть положительным целым числом")
	ErrInvalidBackoffMultiplier     = errors.New("Множитель времени ожидания должен быть положительным целым числом")
)

// Ошибки валидации конфигурации Backoff
var (
	ErrInvalidBaseDelaySeconds         = errors.New("Базовое время задержки должно быть положительным целым числом")
	ErrInvalidMultiplier               = errors.New("Множитель задержки должен быть положительным целым числом")
	ErrInvalidMaxDelaySeconds          = errors.New("Максимальное время задержки должно быть положительным целым числом")
	ErrInvalidJitter                   = errors.New("Разброс задержки должен быть положительным целым числом")
	ErrInvalidMinConnectTimeoutSeconds = errors.New("Минимальное время ожидания соединения должно быть положительным целым числом")
)

// Ошибки валидации конфигурации базы данных
var (
	ErrDbDirIsEmpty      = errors.New("Директория базы данных не может быть пустой")
	ErrDbFileNameIsEmpty = errors.New("Имя файла базы данных не может быть пустым")
)

// Ошибки валидации конфигурации логгера
var (
	ErrLogFileDirectoryIsEmpty = errors.New("Директория для логов не может быть пустой")
	ErrLogFileNameIsEmpty      = errors.New("Имя файла логов не может быть пустым")
	ErrLogFileExtensionIsEmpty = errors.New("Расширение файла логов не может быть пустым")
	ErrInvalidLogLevel         = errors.New("Уровень логирования может быть только DEBUG, INFO, WARN, ERROR")
)

// Ошибки валидации конфигурации TLS
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

// Ошибки валидации конфигурации mDNS сервера
var (
	ErrMDnsServerInstanceNamePostfixIsEmpty = errors.New("Постфикс названия экземпляра сервера mDNS не может быть пустым")
	ErrMDnsServerServiceTypeIsEmpty         = errors.New("Тип сервиса mDNS не может быть пустым")
	ErrMDnsServerDomainIsEmpty              = errors.New("Домен mDNS не может быть пустым")
	ErrMDnsServerPortIsInvalid              = errors.New("Порт mDNS должен быть целым числом в диапозоне от 0 до 65535")
)

// Ошибки валидации конфигурации mDNS браузера
var (
	ErrMDnsBrowserServerServiceTypeIsEmpty         = errors.New("Тип сервиса сервера mDNSв браузере не может быть пустым")
	ErrMDnsBrowserServerInstanceNamePostfixIsEmpty = errors.New("Постфикс названия экземпляра сервера mDNS в браузере не может быть пустым")
	ErrMDnsBrowserServerDomainIsEmpty              = errors.New("Домен сервера mDNS в браузере не может быть пустым")
)

var (
	ErrTempDirIsEmpty = errors.New("Директория для временных файлов не может быть пустой")
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

type LogFileDoesNotExistError struct {
	Err error
}

func (e *LogFileDoesNotExistError) Error() string {
	return "Файл логов не существует: " + e.Err.Error()
}
