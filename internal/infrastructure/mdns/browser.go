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

	logger    *slog.Logger
	loggerCtx context.Context
}

type MDnsNodeNamesBrowserOptions struct {
	ServerServiceType string
	ServerDomain      string
	Interfaces        []string
	Logger            *slog.Logger
}

func NewMDnsNodeNamesBrowser(opts MDnsNodeNamesBrowserOptions) *MDnsNodeNamesBrowser {
	if opts.ServerServiceType == "" ||
		opts.ServerDomain == "" ||
		opts.Interfaces == nil ||
		opts.Logger == nil {
		panic("Все поля MDnsNodeNamesBrowserOptions должны быть заполнены")
	}
	return &MDnsNodeNamesBrowser{
		serverServiceType: opts.ServerServiceType,
		serverDomain:      opts.ServerDomain,
		interfaces:        opts.Interfaces,
		logger:            opts.Logger,
		loggerCtx:         context.Background(),
	}
}

func (b *MDnsNodeNamesBrowser) BrowseNodeNames(ctx context.Context) (chan domain.NodeName, error) {
	b.logger.Info("запуск MDnsNodeNamesBrowser...")

	ifaces, err := shared.GetNetworkInterfaces(b.interfaces)
	if len(ifaces) == 0 {
		b.logger.LogAttrs(
			b.loggerCtx,
			slog.LevelError,
			"Интерфейсы не были переданы для MDnsNodeNamesBrowser",
		)
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
		if err := resolver.Browse(ctx, b.serverServiceType, b.serverDomain, entries); err != nil {
			b.logger.LogAttrs(
				b.loggerCtx,
				slog.LevelError,
				"Ошибка при выполнении метода Browse в MDnsNodeNamesBrowser",
				slog.String("error", err.Error()),
			)
		}
	}()

	nodeNamesChan := make(chan domain.NodeName)

	go func() {
		defer close(nodeNamesChan)
		b.sendToNodeNamesChan(ctx, nodeNamesChan)
	}()

	b.logger.Info("MDnsNodeNamesBrowser успешно запущен")

	return nodeNamesChan, nil
}

func (b *MDnsNodeNamesBrowser) sendToNodeNamesChan(ctx context.Context, nodeChan chan domain.NodeName) {
	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-b.entries:
			if !ok {
				return
			}

			nodeName, err := domain.NewNodeName(entry.Instance)
			if err != nil {
				b.logger.LogAttrs(
					b.loggerCtx,
					slog.LevelError,
					"Ошибка при выполнении метода NewNodeName в MDnsNodeNamesBrowser",
					slog.String("error", err.Error()),
				)
				continue
			}

			nodeChan <- nodeName
		}
	}
}
