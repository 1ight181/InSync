package resolver

import (
	"context"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
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
	loggerCtx := context.Background()
	return &mdnsBuilder{
		resolverIfaces:              opts.ResolverIfaces,
		backgroundListenTimeout:     opts.BackgroundListenTimeout,
		shouldResolveIpv6:           opts.ShouldResolveIpv6,
		shouldDisableResolverOnIdle: opts.ShouldDisableResolverOnIdle,
		shouldReportError:           opts.ShouldReportError,
		logger:                      opts.Logger,
		loggerCtx:                   loggerCtx,
	}
}

func (b *mdnsBuilder) Scheme() string { return scheme }

func (b *mdnsBuilder) Build(target resolver.Target, clientConn resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	targetInfo, err := parseTargetInfo(target)
	if err != nil {
		return nil, err
	}
	var ifaces []net.Interface
	resolverIfaces := b.resolverIfaces
	if resolverIfaces == nil {
		ifaces, err = net.Interfaces()
		if err != nil {
			return nil, err
		}
	} else {
		ifaces, err = shared.GetNetworkInterfacesByName(resolverIfaces)
		if err != nil && len(ifaces) == 0 {
			b.logger.LogAttrs(
				b.loggerCtx,
				slog.LevelWarn,
				"Ни один интерфейс не был успешно загружен, фоллбэк на все интерфейсы",
				slog.String("error", err.Error()),
			)
			ifaces, err = net.Interfaces()
			if err != nil {
				return nil, err
			}
		} else if err != nil && len(ifaces) > 0 {
			successfullyLoadedIfaces := make([]string, 0, len(ifaces))
			for _, iface := range ifaces {
				successfullyLoadedIfaces = append(successfullyLoadedIfaces, iface.Name)
			}
			b.logger.LogAttrs(
				b.loggerCtx,
				slog.LevelWarn,
				"Не все интерфейсы были успешно загружены",
				slog.String("error", err.Error()),
				slog.String("all_ifaces", strings.Join(resolverIfaces, ",")),
				slog.String("successfully_loaded_ifaces", strings.Join(successfullyLoadedIfaces, ",")),
			)
		}
	}

	mDnsResolver, err := zeroconf.NewResolver(zeroconf.SelectIfaces(ifaces))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	mdnsResolverOptions := mdnsResolverOptions{
		TargetInfo:                  targetInfo,
		ClientConn:                  clientConn,
		Ctx:                         ctx,
		Cancel:                      cancel,
		Resolver:                    mDnsResolver,
		BackgroundListenTimeout:     b.backgroundListenTimeout,
		ShouldResolveIpv6:           b.shouldResolveIpv6,
		ShouldDisableResolverOnIdle: b.shouldDisableResolverOnIdle,
		ShouldReportError:           b.shouldReportError,
		Logger:                      b.logger,
	}

	mdnsResolver := newMdnsResolver(mdnsResolverOptions)
	mdnsResolver.Start()

	return mdnsResolver, nil
}
