package mdns

import (
	"context"
	"log/slog"

	shared "insync/internal/shared"

	"github.com/grandcat/zeroconf"
)

type MDnsNodeNamesBrowser struct {
	serviceType string
	domain      string

	interfaces []string

	entries chan *zeroconf.ServiceEntry

	logger *slog.Logger
	ctx    context.Context
}

type MDnsNodeNamesBrowserOptions struct {
	ServiceType string
	Domain      string
	Interfaces  []string
	Logger      *slog.Logger
	Ctx         context.Context
}

func NewMDnsNodeNamesBrowser(opts MDnsNodeNamesBrowserOptions) *MDnsNodeNamesBrowser {
	if opts.ServiceType == "" ||
		opts.Domain == "" ||
		opts.Interfaces == nil ||
		opts.Logger == nil ||
		opts.Ctx == nil {
		panic("Все поля MDnsResolverOptions должны быть заполнены")
	}
	return &MDnsNodeNamesBrowser{
		serviceType: opts.ServiceType,
		domain:      opts.Domain,
		interfaces:  opts.Interfaces,
		logger:      opts.Logger,
		ctx:         opts.Ctx,
	}
}

func (b *MDnsNodeNamesBrowser) BrowseNodeNames() (chan string, error) {
	b.logger.Info("запуск MDnsResolver...")

	ifaces, err := shared.GetNetworkInterfacesByName(b.interfaces)
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

	nodeNamesChan := make(chan string)

	go func() {
		defer close(nodeNamesChan)
		b.sendToNodeNamesChan(nodeNamesChan)
	}()

	b.logger.Info("MDnsResolver успешно запущен")

	return nodeNamesChan, nil
}

func (b *MDnsNodeNamesBrowser) sendToNodeNamesChan(nodeChan chan string) {
	for {
		select {
		case <-b.ctx.Done():
			return
		case entry, ok := <-b.entries:
			if !ok {
				return
			}

			nodeChan <- entry.Instance
		}
	}
}
