package resolver

import (
	"context"
	"log/slog"
	"net"
	"strings"

	"google.golang.org/grpc/resolver"

	shared "insync/internal/shared"
)

const (
	scheme = "mdns"
)

func init() {
	resolver.Register(&mdnsBuilder{})
}

type mdnsBuilder struct {
	resolverIfaces    []string
	shouldResolveIpv6 bool
	shouldReportError bool
	logger            *slog.Logger
	loggerCtx         context.Context
}

type BuilderOptions struct {
	ResolverIfaces    []string
	ShouldResolveIpv6 bool
	ShouldReportError bool
	Logger            *slog.Logger
}

func NewBuilder(opts BuilderOptions) *mdnsBuilder {
	return &mdnsBuilder{
		resolverIfaces:    opts.ResolverIfaces,
		shouldResolveIpv6: opts.ShouldResolveIpv6,
		shouldReportError: opts.ShouldReportError,
		logger:            opts.Logger,
		loggerCtx:         context.Background(),
	}
}

func (b *mdnsBuilder) Scheme() string { return scheme }

func (b *mdnsBuilder) Build(target resolver.Target, clientConn resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	targetInfo, err := parseTargetInfo(target)
	if err != nil {
		return nil, err
	}

	// Загружаем интерфейсы
	var ifaces []net.Interface
	if len(b.resolverIfaces) == 0 {
		ifaces, err = net.Interfaces()
		if err != nil {
			return nil, err
		}
	} else {
		ifaces, err = shared.GetNetworkInterfaces(b.resolverIfaces)
		if err != nil || len(ifaces) == 0 {
			b.logger.LogAttrs(
				b.loggerCtx,
				slog.LevelWarn,
				"Не удалось загрузить указанные интерфейсы, фоллбэк на все интерфейсы",
				slog.String("error", err.Error()),
				slog.Any("requested_ifaces", b.resolverIfaces),
			)
			ifaces, err = net.Interfaces()
			if err != nil {
				return nil, err
			}
		} else if len(ifaces) > 0 {
			loaded := make([]string, 0, len(ifaces))
			for _, i := range ifaces {
				loaded = append(loaded, i.Name)
			}
			b.logger.LogAttrs(
				b.loggerCtx,
				slog.LevelWarn,
				"Не все интерфейсы удалось загрузить",
				slog.String("requested", strings.Join(b.resolverIfaces, ",")),
				slog.String("loaded", strings.Join(loaded, ",")),
			)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	mdnsResolver := newMDnsResolver(mdnsResolverOptions{
		TargetInfo:        targetInfo,
		ClientConn:        clientConn,
		Ctx:               ctx,
		Cancel:            cancel,
		Interfaces:        ifaces,
		ShouldResolveIpv6: b.shouldResolveIpv6,
		ShouldReportError: b.shouldReportError,
		Logger:            b.logger,
	})

	mdnsResolver.Start()

	return mdnsResolver, nil
}
