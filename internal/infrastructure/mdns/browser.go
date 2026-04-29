package mdns

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"insync/internal/domain"
	shared "insync/internal/shared"

	"github.com/hashicorp/mdns"
)

type MDnsNodeNamesBrowser struct {
	serverServiceType string
	serverDomain      string

	interfaces []string

	logger        *slog.Logger
	loggerCtx     context.Context
	seenNodeNames map[domain.NodeName]struct{}
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
		seenNodeNames:     make(map[domain.NodeName]struct{}),
	}
}

func (b *MDnsNodeNamesBrowser) BrowseNodeNames(ctx context.Context) (chan domain.NodeName, error) {
	b.logger.Info("запуск MDnsNodeNamesBrowser...")

	ifaces, err := shared.GetNetworkInterfaces(b.interfaces)
	if err != nil || len(ifaces) == 0 {
		b.logger.LogAttrs(
			b.loggerCtx,
			slog.LevelError,
			"Интерфейсы не были переданы или не найдены для MDnsNodeNamesBrowser",
			slog.Any("interfaces", b.interfaces),
			slog.String("error", fmt.Sprintf("%v", err)),
		)
		if err == nil {
			err = fmt.Errorf("no network interfaces found")
		}
		return nil, err
	}

	b.logger.LogAttrs(
		b.loggerCtx,
		slog.LevelDebug,
		"Загружены интерфейсы для MDnsNodeNamesBrowser",
		slog.Any("interfaces", ifaces),
	)

	nodeNamesChan := make(chan domain.NodeName)

	go func() {
		defer close(nodeNamesChan)
		b.seenNodeNames = make(map[domain.NodeName]struct{})
		b.discoveryLoop(ctx, ifaces, nodeNamesChan)
	}()

	b.logger.Info("MDnsNodeNamesBrowser успешно запущен")
	return nodeNamesChan, nil
}

func (b *MDnsNodeNamesBrowser) discoveryLoop(ctx context.Context, ifaces []net.Interface, nodeChan chan domain.NodeName) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.queryAllInterfaces(ctx, ifaces, nodeChan)
		}
	}
}

func (b *MDnsNodeNamesBrowser) queryAllInterfaces(ctx context.Context, ifaces []net.Interface, nodeChan chan domain.NodeName) {
	for _, iface := range ifaces {
		select {
		case <-ctx.Done():
			return
		default:
		}

		entriesCh := make(chan *mdns.ServiceEntry, 16)

		params := &mdns.QueryParam{
			Service:             b.serverServiceType,
			Domain:              b.serverDomain,
			Timeout:             1500 * time.Millisecond,
			Interface:           &iface,
			Entries:             entriesCh,
			WantUnicastResponse: false,
		}

		go func(ifaceName string) {
			if err := mdns.Query(params); err != nil {
				b.logger.LogAttrs(
					b.loggerCtx,
					slog.LevelWarn,
					"Ошибка mDNS Query",
					slog.String("interface", ifaceName),
					slog.String("error", err.Error()),
				)
			}
		}(iface.Name)

		timeout := time.After(2 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-timeout:
				goto nextInterface
			case entry, ok := <-entriesCh:
				if !ok {
					goto nextInterface
				}
				b.processEntry(entry, nodeChan)
			}
		}
	nextInterface:
	}
}

func (b *MDnsNodeNamesBrowser) processEntry(entry *mdns.ServiceEntry, nodeChan chan domain.NodeName) {
	if entry == nil || entry.Name == "" {
		return
	}

	nodeName, err := domain.NewNodeName(entry.Name)
	if err != nil {
		b.logger.LogAttrs(
			b.loggerCtx,
			slog.LevelError,
			"Ошибка при выполнении метода NewNodeName в MDnsNodeNamesBrowser",
			slog.String("instance", entry.Name),
			slog.String("error", err.Error()),
		)
		return
	}

	if _, exists := b.seenNodeNames[nodeName]; exists {
		return
	}
	b.seenNodeNames[nodeName] = struct{}{}

	select {
	case nodeChan <- nodeName:
	default:
		b.logger.LogAttrs(
			b.loggerCtx,
			slog.LevelWarn,
			"Канал nodeNamesChan переполнен, пропускаем NodeName",
			slog.String("node", nodeName.String()),
		)
	}
}
