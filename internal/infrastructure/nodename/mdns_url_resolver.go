package nodename

import (
	"fmt"
	"insync/internal/domain"
)

type MDnsUrlResolver struct {
	mDnsServerServiceType string
	mDnsServerDomain      string
}

type MDnsUrlResolverOptions struct {
	MDnsServerServiceType string
	MDnsServerDomain      string
}

func NewMDnsUrlResolver(opts MDnsUrlResolverOptions) *MDnsUrlResolver {
	if opts.MDnsServerDomain == "" || opts.MDnsServerServiceType == "" {
		panic("Все поля MDnsUrlResolverOptions должны быть заполнены")
	}
	return &MDnsUrlResolver{
		mDnsServerServiceType: opts.MDnsServerServiceType,
		mDnsServerDomain:      opts.MDnsServerDomain,
	}
}

func (r *MDnsUrlResolver) Resolve(nodeName domain.NodeName) string {
	return fmt.Sprintf("%s.%s.%s", r.mDnsServerServiceType, nodeName, r.mDnsServerDomain)
}
