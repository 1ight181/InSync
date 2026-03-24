package resolver

import (
	"context"
	"net"
	"time"

	"github.com/grandcat/zeroconf"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"

	shared "insync/internal/shared"
)

var logger = grpclog.Component("mdns")

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
}

type BuilderOptions struct {
	ResolverIfaces              []string
	BackgroundListenTimeout     time.Duration
	ShouldResolveIpv6           bool
	ShouldDisableResolverOnIdle bool
	ShouldReportError           bool
}

func NewBuilder(opts BuilderOptions) resolver.Builder {
	if opts.BackgroundListenTimeout == 0 {
		opts.BackgroundListenTimeout = defaultBackgroundListenTimeout
	}
	return &mdnsBuilder{
		resolverIfaces:              opts.ResolverIfaces,
		backgroundListenTimeout:     opts.BackgroundListenTimeout,
		shouldResolveIpv6:           opts.ShouldResolveIpv6,
		shouldDisableResolverOnIdle: opts.ShouldDisableResolverOnIdle,
		shouldReportError:           opts.ShouldReportError,
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
			logger.Warningf("Some or all of the interfaces were not successfully parsed, falling back to all interfaces: %v", err)
			ifaces, err = net.Interfaces()
			if err != nil {
				return nil, err
			}
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
	}

	mdnsResolver := newMdnsResolver(mdnsResolverOptions)
	mdnsResolver.Start()

	return mdnsResolver, nil
}
