package mdns

import (
	"context"
	"log/slog"
	"sync/atomic"

	ifaces "insync/internal/interfaces"

	"github.com/grandcat/zeroconf"
)

type MDnsServer struct {
	instanceName string
	serviceType  string
	domain       string
	port         int
	interfaces   []string

	logger *slog.Logger
	ctx    context.Context

	server    *zeroconf.Server
	isStarted atomic.Bool
}

type MDnsServerOptions struct {
	InstanceName string
	ServiceType  string
	Domain       string
	Port         int
	Interfaces   []string
	Logger       *slog.Logger
	Ctx          context.Context
}

func NewMDnsServer(opts MDnsServerOptions) ifaces.IMDnsServer {
	if opts.InstanceName == "" ||
		opts.ServiceType == "" ||
		opts.Domain == "" ||
		opts.Port == 0 ||
		opts.Interfaces == nil ||
		opts.Logger == nil ||
		opts.Ctx == nil {
		panic("Все поля MDnsServerOptions должны быть заполнены")
	}
	return &MDnsServer{
		instanceName: opts.InstanceName,
		serviceType:  opts.ServiceType,
		domain:       opts.Domain,
		port:         opts.Port,
		interfaces:   opts.Interfaces,
		logger:       opts.Logger,
		ctx:          opts.Ctx,
	}
}

func (ms *MDnsServer) Start() error {
	if ms.isStarted.Swap(true) {
		ms.logger.Warn("Попытка запустить mDNS, который уже запущен")
		return ErrServerAlreadyStarted
	}

	ms.logger.Info("Запуск mDNS сервера...")
	interfaces, err := getNetworkInterfacesByName(ms.interfaces)
	if err != nil {
		ms.isStarted.Store(false)
		return err
	}

	ms.logger.LogAttrs(
		ms.ctx,
		slog.LevelDebug,
		"Загружены интерфейсы для MDnsServer",
		slog.Any("interfaces", interfaces),
	)

	server, err := zeroconf.Register(
		ms.instanceName,
		ms.serviceType,
		ms.domain,
		ms.port,
		nil,
		interfaces,
	)
	if err != nil {
		ms.isStarted.Store(false)
		return err
	}

	ms.server = server

	ms.logger.Info("mDNS сервер успешно запущен")

	return nil
}

func (ms *MDnsServer) Stop() error {
	if !ms.isStarted.Swap(false) {
		ms.logger.Warn("Попытка остановить mDNS, который не запущен")
		return ErrServerAlreadyStopped
	}

	ms.server.Shutdown()
	ms.logger.Info("mDNS сервер успешно остановлен")

	return nil
}
