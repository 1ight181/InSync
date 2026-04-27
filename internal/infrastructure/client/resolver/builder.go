package resolver

import (
	"context"
	"log/slog"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc/resolver"

	shared "insync/internal/shared"
)

const (
	scheme                         = "mdns"
	defaultBackgroundListenTimeout = 20 * time.Second
)

func init() {
	resolver.Register(&mdnsBuilder{})
}

type mdnsBuilder struct {
	resolverIfaces              []string
	backgroundListenTimeout     time.Duration
	shouldResolveIpv6           bool
	shouldDisableResolverOnIdle bool
	shouldReportError           bool
	logger                      *slog.Logger
	loggerCtx                   context.Context
}

type BuilderOptions struct {
	ResolverIfaces              []string
	BackgroundListenTimeout     time.Duration
	ShouldResolveIpv6           bool
	ShouldDisableResolverOnIdle bool
	ShouldReportError           bool
	Logger                      *slog.Logger
}

func NewBuilder(opts BuilderOptions) *mdnsBuilder {
	if opts.BackgroundListenTimeout == 0 {
		opts.BackgroundListenTimeout = defaultBackgroundListenTimeout
	}
	return &mdnsBuilder{
		resolverIfaces:              opts.ResolverIfaces,
		backgroundListenTimeout:     opts.BackgroundListenTimeout,
		shouldResolveIpv6:           opts.ShouldResolveIpv6,
		shouldDisableResolverOnIdle: opts.ShouldDisableResolverOnIdle,
		shouldReportError:           opts.ShouldReportError,
		logger:                      opts.Logger,
		loggerCtx:                   context.Background(),
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
		TargetInfo:                  targetInfo,
		ClientConn:                  clientConn,
		Ctx:                         ctx,
		Cancel:                      cancel,
		Interfaces:                  ifaces, // ← передаём список интерфейсов
		BackgroundListenTimeout:     b.backgroundListenTimeout,
		ShouldResolveIpv6:           b.shouldResolveIpv6,
		ShouldDisableResolverOnIdle: b.shouldDisableResolverOnIdle,
		ShouldReportError:           b.shouldReportError,
		Logger:                      b.logger,
	})

	mdnsResolver.Start()

	return mdnsResolver, nil
}
