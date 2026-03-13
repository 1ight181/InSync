package mdns

import (
	"context"
	"log/slog"

	mdnsfaces "insync/internal/mdns/interfaces"
	models "insync/internal/mdns/models"

	"github.com/grandcat/zeroconf"
)

type MDnsResolver struct {
	serviceType string
	domain      string

	interfaces []string

	entries chan *zeroconf.ServiceEntry

	logger *slog.Logger
	ctx    context.Context
}

type MDnsResolverOptions struct {
	ServiceType string
	Domain      string
	Interfaces  []string
	Logger      *slog.Logger
	Ctx         context.Context
}

func NewMDnsResolver(opts MDnsResolverOptions) mdnsfaces.IResolver {
	if opts.ServiceType == "" ||
		opts.Domain == "" ||
		opts.Interfaces == nil ||
		opts.Logger == nil ||
		opts.Ctx == nil {
		panic("Все поля MDnsResolverOptions должны быть заполнены")
	}
	return &MDnsResolver{
		serviceType: opts.ServiceType,
		domain:      opts.Domain,
		interfaces:  opts.Interfaces,
		logger:      opts.Logger,
		ctx:         opts.Ctx,
	}
}

// StartBrowsing возвращает канал, куда отдается словарь с адресом в формате Ip:Port
func (r *MDnsResolver) Browse() (chan models.ResolvedAddresses, error) {
	r.logger.Info("запуск MDnsResolver...")

	ifaces, err := getNetworkInterfacesByName(r.interfaces)
	if err != nil {
		return nil, err
	}
	resolver, err := zeroconf.NewResolver(
		zeroconf.SelectIfaces(ifaces),
	)
	if err != nil {
		return nil, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	r.entries = entries

	go func() {
		if err := resolver.Browse(r.ctx, r.serviceType, r.domain, entries); err != nil {
			r.logger.LogAttrs(
				r.ctx,
				slog.LevelError,
				"Ошибка при выполнении метода Browse в MDnsResolver",
				slog.String("error", err.Error()),
			)
		}
	}()

	addressesChan := make(chan models.ResolvedAddresses)

	go func() {
		defer close(addressesChan)
		for {
			select {
			case <-r.ctx.Done():
				return
			case entry, ok := <-entries:
				if !ok {
					return
				}
				addressesChan <- models.ResolvedAddresses{
					Name: entry.Instance,
					Ip:   entry.AddrIPv4,
					Port: entry.Port,
				}
			}
		}
	}()

	r.logger.Info("MDnsResolver успешно запущен")

	return addressesChan, nil
}
