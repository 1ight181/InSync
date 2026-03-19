package mdns

import (
	"context"
	"log/slog"

	"insync/internal/domain"
	ifaces "insync/internal/interfaces"

	"github.com/grandcat/zeroconf"
)

type MDnsBrowser struct {
	serviceType string
	domain      string

	interfaces []string

	entries chan *zeroconf.ServiceEntry

	logger *slog.Logger
	ctx    context.Context
}

type MDnsBrowserOptions struct {
	ServiceType string
	Domain      string
	Interfaces  []string
	Logger      *slog.Logger
	Ctx         context.Context
}

func NewMDnsBrowser(opts MDnsBrowserOptions) ifaces.IMdnsBrowser {
	if opts.ServiceType == "" ||
		opts.Domain == "" ||
		opts.Interfaces == nil ||
		opts.Logger == nil ||
		opts.Ctx == nil {
		panic("Все поля MDnsResolverOptions должны быть заполнены")
	}
	return &MDnsBrowser{
		serviceType: opts.ServiceType,
		domain:      opts.Domain,
		interfaces:  opts.Interfaces,
		logger:      opts.Logger,
		ctx:         opts.Ctx,
	}
}

func (b *MDnsBrowser) Browse() (chan domain.Node, error) {
	b.logger.Info("запуск MDnsResolver...")

	ifaces, err := getNetworkInterfacesByName(b.interfaces)
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
	b.entries = entries

	go func() {
		if err := resolver.Browse(b.ctx, b.serviceType, b.domain, entries); err != nil {
			b.logger.LogAttrs(
				b.ctx,
				slog.LevelError,
				"Ошибка при выполнении метода Browse в MDnsResolver",
				slog.String("error", err.Error()),
			)
		}
	}()

	nodeChan := make(chan domain.Node)

	go func() {
		defer close(nodeChan)
		b.sendToNodeChan(nodeChan)
	}()

	b.logger.Info("MDnsResolver успешно запущен")

	return nodeChan, nil
}

func (b *MDnsBrowser) sendToNodeChan(nodeChan chan domain.Node) {
	for {
		select {
		case <-b.ctx.Done():
			return
		case entry, ok := <-b.entries:
			if !ok {
				return
			}

			addresses := make([]domain.Address, 0, len(entry.AddrIPv4)+len(entry.AddrIPv6))
			for _, ipv4 := range entry.AddrIPv4 {
				addresses = append(addresses, domain.Address{
					Ip:   ipv4.String(),
					Port: entry.Port,
				})
			}

			for _, ipv6 := range entry.AddrIPv6 {
				addresses = append(addresses, domain.Address{
					Ip:   ipv6.String(),
					Port: entry.Port,
				})
			}

			nodeChan <- domain.Node{
				Name:      entry.Instance,
				Addresses: addresses,
			}
		}
	}
}
