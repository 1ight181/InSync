package resolver

import (
	"context"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
	"google.golang.org/grpc/resolver"
)

type mdnsResolver struct {
	targetInfo targetInfo
	clientConn resolver.ClientConn
	ctx        context.Context
	cancel     context.CancelFunc

	interfaces []net.Interface

	shouldResolveIpv6 bool

	shouldReportError bool

	logger    *slog.Logger
	loggerCtx context.Context

	resolveNowChan chan struct{}
	entries        chan *mdns.ServiceEntry
	wg             sync.WaitGroup

	lastAddresses map[string]struct{}
}

type mdnsResolverOptions struct {
	TargetInfo targetInfo
	ClientConn resolver.ClientConn
	Ctx        context.Context
	Cancel     context.CancelFunc

	Interfaces []net.Interface

	ShouldResolveIpv6           bool
	ShouldDisableResolverOnIdle bool
	BackgroundListenTimeout     time.Duration
	ShouldReportError           bool

	Logger *slog.Logger
}

func newMDnsResolver(opts mdnsResolverOptions) *mdnsResolver {
	return &mdnsResolver{
		targetInfo:        opts.TargetInfo,
		clientConn:        opts.ClientConn,
		ctx:               opts.Ctx,
		cancel:            opts.Cancel,
		interfaces:        opts.Interfaces,
		shouldResolveIpv6: opts.ShouldResolveIpv6,
		shouldReportError: opts.ShouldReportError,
		logger:            opts.Logger,
		loggerCtx:         context.Background(),
		resolveNowChan:    make(chan struct{}, 1),
		entries:           make(chan *mdns.ServiceEntry, 32),
		lastAddresses:     make(map[string]struct{}),
	}
}

func (r *mdnsResolver) Close() {
	r.cancel()
	r.wg.Wait()
	close(r.resolveNowChan)
	close(r.entries)
}

func (r *mdnsResolver) ResolveNow(options resolver.ResolveNowOptions) {
	r.logger.LogAttrs(r.loggerCtx, slog.LevelDebug, "получен запрос на ResolveNow")
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

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.resolveNowChan:
			r.performLookup()
		}
	}
}

func (r *mdnsResolver) performLookup() {
	for _, iface := range r.interfaces {
		ctx := r.ctx
		if ctx.Err() != nil {
			return
		}

		entriesCh := make(chan *mdns.ServiceEntry, 16)

		params := &mdns.QueryParam{
			Service:   r.targetInfo.serviceType,
			Domain:    r.targetInfo.domain,
			Timeout:   2500 * time.Millisecond,
			Interface: &iface,
			Entries:   entriesCh,
		}

		r.wg.Add(1)
		go func(ifaceName string) {
			defer r.wg.Done()
			if err := mdns.Query(params); err != nil && r.shouldReportError {
				r.logger.LogAttrs(r.loggerCtx, slog.LevelWarn,
					"mDNS Query ошибка",
					slog.String("interface", ifaceName),
					slog.String("error", err.Error()),
				)
				r.clientConn.ReportError(err)
			}
			close(entriesCh)
		}(iface.Name)

		for entry := range entriesCh {
			if entry == nil {
				continue
			}

			if r.targetInfo.instanceName != "" && !strings.HasPrefix(entry.Name, r.targetInfo.instanceName) {
				continue
			}

			select {
			case r.entries <- entry:
			case <-r.ctx.Done():
				return
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
				continue
			}

			addrs := r.buildAddresses(entry)
			if len(addrs) == 0 {
				continue
			}

			if r.isSameAddresses(addrs) {
				continue
			}

			r.updateLastAddresses(addrs)

			logAddrs := make([]string, 0, len(addrs))
			for _, a := range addrs {
				logAddrs = append(logAddrs, a.Addr)
			}

			r.logger.LogAttrs(
				r.loggerCtx,
				slog.LevelDebug,
				"Найдены адреса через mDNS",
				slog.String("addrs", strings.Join(logAddrs, ", ")),
				slog.String("interface", entry.Host),
				slog.String("mdns_domain", r.targetInfo.domain),
				slog.String("mdns_service_name", r.targetInfo.serviceType),
				slog.String("mdns_instance_name", r.targetInfo.instanceName),
			)

			if err := r.clientConn.UpdateState(resolver.State{Addresses: addrs}); err != nil && r.shouldReportError {
				r.clientConn.ReportError(err)
			}
		}
	}
}

func (r *mdnsResolver) buildAddresses(entry *mdns.ServiceEntry) []resolver.Address {
	var addrs []resolver.Address

	// IPv4
	if entry.AddrV4 != nil {
		addrStr := net.JoinHostPort(entry.AddrV4.String(), strconv.Itoa(entry.Port))
		addrs = append(addrs, resolver.Address{Addr: addrStr})
	}

	// IPv6
	if r.shouldResolveIpv6 {
		if entry.AddrV6IPAddr.IP != nil {
			addrStr := net.JoinHostPort(entry.AddrV6.String(), strconv.Itoa(entry.Port))
			addrs = append(addrs, resolver.Address{Addr: addrStr})
		}
	}

	return addrs
}

func (r *mdnsResolver) isSameAddresses(current []resolver.Address) bool {
	if len(current) != len(r.lastAddresses) {
		return false
	}
	for _, a := range current {
		if _, ok := r.lastAddresses[a.Addr]; !ok {
			return false
		}
	}
	return true
}

func (r *mdnsResolver) updateLastAddresses(addrs []resolver.Address) {
	clear(r.lastAddresses)
	for _, a := range addrs {
		r.lastAddresses[a.Addr] = struct{}{}
	}
}
