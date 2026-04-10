package mdns

import (
	"context"
	"log/slog"

	"insync/internal/domain"
	shared "insync/internal/shared"

	"github.com/grandcat/zeroconf"
)

type MDnsNodeNamesBrowser struct {
	serverServiceType string
	serverDomain      string

	interfaces []string

	entries chan *zeroconf.ServiceEntry

	logger *slog.Logger
	ctx    context.Context
}

type MDnsNodeNamesBrowserOptions struct {
	ServerServiceType string
	ServerDomain      string
	Interfaces        []string
	Logger            *slog.Logger
	Ctx               context.Context
}

func NewMDnsNodeNamesBrowser(opts MDnsNodeNamesBrowserOptions) *MDnsNodeNamesBrowser {
	if opts.ServerServiceType == "" ||
		opts.ServerDomain == "" ||
		opts.Interfaces == nil ||
		opts.Logger == nil ||
		opts.Ctx == nil {
		panic("Все поля MDnsResolverOptions должны быть заполнены")
	}
	return &MDnsNodeNamesBrowser{
		serverServiceType: opts.ServerServiceType,
		serverDomain:      opts.ServerDomain,
		interfaces:        opts.Interfaces,
		logger:            opts.Logger,
		ctx:               opts.Ctx,
	}
}

func (b *MDnsNodeNamesBrowser) BrowseNodeNames() (chan domain.NodeName, error) {
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
		if err := resolver.Browse(b.ctx, b.serverServiceType, b.serverDomain, entries); err != nil {
			b.logger.LogAttrs(
				b.ctx,
				slog.LevelError,
				"Ошибка при выполнении метода Browse в MDnsResolver",
				slog.String("error", err.Error()),
			)
		}
	}()

	nodeNamesChan := make(chan domain.NodeName)

	go func() {
		defer close(nodeNamesChan)
		b.sendToNodeNamesChan(nodeNamesChan)
	}()

	b.logger.Info("MDnsResolver успешно запущен")

	return nodeNamesChan, nil
}

func (b *MDnsNodeNamesBrowser) sendToNodeNamesChan(nodeChan chan domain.NodeName) {
	for {
		select {
		case <-b.ctx.Done():
			return
		case entry, ok := <-b.entries:
			if !ok {
				return
			}

			nodeName, err := domain.NewNodeName(entry.Instance)
			if err != nil {
				b.logger.LogAttrs(
					b.ctx,
					slog.LevelError,
					"Ошибка при выполнении метода NewNodeName в MDnsResolver",
					slog.String("error", err.Error()),
				)
				continue
			}

			nodeChan <- nodeName
		}
	}
}
