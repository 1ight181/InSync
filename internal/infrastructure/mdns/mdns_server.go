package mdns

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	shared "insync/internal/shared"

	"github.com/hashicorp/mdns"
)

type MDnsServer struct {
	instanceName string
	serviceType  string
	domain       string
	port         int
	interfaces   []string

	logger    *slog.Logger
	loggerCtx context.Context

	server    *mdns.Server
	isStarted atomic.Bool
}

type MDnsServerOptions struct {
	InstanceName string
	ServiceType  string
	Domain       string
	Port         int
	Interfaces   []string
	Logger       *slog.Logger
}

func NewMDnsServer(opts MDnsServerOptions) *MDnsServer {
	if opts.InstanceName == "" ||
		opts.ServiceType == "" ||
		opts.Domain == "" ||
		opts.Port == 0 ||
		opts.Interfaces == nil ||
		opts.Logger == nil {
		panic("Все поля MDnsServerOptions должны быть заполнены")
	}

	return &MDnsServer{
		instanceName: opts.InstanceName,
		serviceType:  opts.ServiceType,
		domain:       opts.Domain,
		port:         opts.Port,
		interfaces:   opts.Interfaces,
		logger:       opts.Logger,
		loggerCtx:    context.Background(),
	}
}

func (ms *MDnsServer) Start() error {
	if ms.isStarted.Swap(true) {
		ms.logger.Warn("Попытка запустить mDNS, который уже запущен")
		return ErrServerAlreadyStarted
	}

	ms.logger.Info("Запуск mDNS сервера...")

	// Проверяем и логируем интерфейсы (для совместимости с твоим shared кодом)
	ifaces, err := shared.GetNetworkInterfaces(ms.interfaces)
	if err != nil || len(ifaces) == 0 {
		ms.logger.LogAttrs(
			ms.loggerCtx,
			slog.LevelError,
			"Интерфейсы не были переданы или не найдены для MDnsServer",
			slog.Any("interfaces", ms.interfaces),
			slog.String("error", fmt.Sprintf("%v", err)),
		)

		ms.isStarted.Store(false)
		if err == nil {
			err = fmt.Errorf("no network interfaces found")
		}
		return err
	}

	ms.logger.LogAttrs(
		ms.loggerCtx,
		slog.LevelDebug,
		"Загружены интерфейсы для MDnsServer",
		slog.Any("interfaces", ifaces),
	)

	// Создаём описание сервиса
	service, err := mdns.NewMDNSService(
		ms.instanceName, // Instance name
		ms.serviceType,  // Service type (_myapp._tcp)
		ms.domain,       // Domain (обычно "local.")
		"",              // Host name (пусто = берётся автоматически)
		ms.port,
		nil, // IPs: nil = все адреса интерфейса
		nil, // TXT records (можно добавить []string{...} при необходимости)
	)
	if err != nil {
		ms.isStarted.Store(false)
		return fmt.Errorf("failed to create MDNSService: %w", err)
	}

	// Запускаем mDNS сервер
	server, err := mdns.NewServer(&mdns.Config{
		Zone: service,
		// Iface: nil — слушает на всех интерфейсах (рекомендуется для большинства случаев)
	})
	if err != nil {
		ms.isStarted.Store(false)
		return fmt.Errorf("failed to create mDNS server: %w", err)
	}

	ms.server = server

	ms.logger.Info("mDNS сервер успешно запущен",
		slog.String("instance", ms.instanceName),
		slog.String("service_type", ms.serviceType),
		slog.Int("port", ms.port),
	)

	return nil
}

func (ms *MDnsServer) Stop() error {
	if !ms.isStarted.Swap(false) {
		ms.logger.Warn("Попытка остановить mDNS, который не запущен")
		return ErrServerAlreadyStopped
	}

	if ms.server != nil {
		ms.server.Shutdown()
		ms.server = nil
	}

	ms.logger.Info("mDNS сервер успешно остановлен")
	return nil
}
