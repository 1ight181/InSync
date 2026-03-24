package resolver

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	"google.golang.org/grpc/resolver"
)

type mdnsResolver struct {
	targetInfo targetInfo
	clientConn resolver.ClientConn
	ctx        context.Context
	cancel     context.CancelFunc
	resolver   *zeroconf.Resolver

	shouldResolveIpv6 bool

	shouldDisableResolverOnIdle bool
	backgroundListenTimeout     time.Duration
	shouldReportError           bool

	resolveNowChan chan struct{}
	entries        chan *zeroconf.ServiceEntry
	wg             sync.WaitGroup
}

type mdnsResolverOptions struct {
	TargetInfo targetInfo
	ClientConn resolver.ClientConn
	Ctx        context.Context
	Cancel     context.CancelFunc
	Resolver   *zeroconf.Resolver

	ShouldResolveIpv6 bool

	ShouldDisableResolverOnIdle bool
	BackgroundListenTimeout     time.Duration
	ShouldReportError           bool
}

func newMdnsResolver(opts mdnsResolverOptions) *mdnsResolver {
	resolveNowChan := make(chan struct{}, 1)
	entries := make(chan *zeroconf.ServiceEntry)
	return &mdnsResolver{
		targetInfo: opts.TargetInfo,
		clientConn: opts.ClientConn,
		ctx:        opts.Ctx,
		cancel:     opts.Cancel,
		resolver:   opts.Resolver,

		backgroundListenTimeout: opts.BackgroundListenTimeout,

		shouldResolveIpv6:           opts.ShouldResolveIpv6,
		shouldDisableResolverOnIdle: opts.ShouldDisableResolverOnIdle,
		shouldReportError:           opts.ShouldReportError,

		resolveNowChan: resolveNowChan,
		entries:        entries,
	}
}

func (r *mdnsResolver) Close() {
	r.cancel()
	r.wg.Wait()
	close(r.entries)
	close(r.resolveNowChan)
}

func (r *mdnsResolver) ResolveNow(options resolver.ResolveNowOptions) {
	select {
	case r.resolveNowChan <- struct{}{}:
	default:
	}
}

func (r *mdnsResolver) Start() {
	r.wg.Add(2)
	go r.lookup()
	go r.watcher()
	r.ResolveNow(resolver.ResolveNowOptions{})
}

func (r *mdnsResolver) lookup() {
	defer r.wg.Done()
	parentCtx := r.ctx
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.resolveNowChan:
			lookupCtx := parentCtx
			var lookupCancel context.CancelFunc
			if r.shouldDisableResolverOnIdle {
				lookupCtx, lookupCancel = context.WithTimeout(parentCtx, r.backgroundListenTimeout)
			}
			err := r.resolver.Lookup(
				lookupCtx,
				r.targetInfo.instanceName,
				r.targetInfo.serviceName,
				r.targetInfo.domain,
				r.entries,
			)
			if err != nil && r.shouldReportError {
				r.clientConn.ReportError(err)
			}
			select {
			case <-parentCtx.Done():
				if lookupCancel != nil {
					lookupCancel()
				}
				return

			case <-lookupCtx.Done():
				if lookupCancel != nil {
					lookupCancel()
				}
				continue
			}

		}
	}
}

func (r *mdnsResolver) watcher() {
	defer r.wg.Done()
	for {
		select {
		case <-r.ctx.Done():
			return
		case entry, ok := <-r.entries:
			if !ok {
				return
			}
			if entry != nil {
				addrs := make([]resolver.Address, 0)

				ipv4 := entry.AddrIPv4
				for _, ip := range ipv4 {
					addr := net.JoinHostPort(ip.String(), strconv.Itoa(entry.Port))
					addrs = append(addrs, resolver.Address{Addr: addr})
				}

				if r.shouldResolveIpv6 {
					ipv6 := entry.AddrIPv6
					for _, ip := range ipv6 {
						addr := net.JoinHostPort(ip.String(), strconv.Itoa(entry.Port))
						addrs = append(addrs, resolver.Address{Addr: addr})
					}
				}

				state := resolver.State{Addresses: addrs}
				err := r.clientConn.UpdateState(state)
				if err != nil && r.shouldReportError {
					r.clientConn.ReportError(err)
				}
			}
		}
	}
}
