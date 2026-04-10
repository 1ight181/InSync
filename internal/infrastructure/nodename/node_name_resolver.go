package nodename

import (
	"fmt"
	"insync/internal/domain"
)

type NodeNameResolver struct {
	mDnsServerServiceType string
	mDnsServerDomain      string
}

type NodeNameResolverOptions struct {
	MDnsServerServiceType string
	MDnsServerDomain      string
}

func NewNodeNameResolver(opts NodeNameResolverOptions) *NodeNameResolver {
	if opts.MDnsServerDomain == "" || opts.MDnsServerServiceType == "" {
		panic("Все поля NodeNameResolverOptions должны быть заполнены")
	}
	return &NodeNameResolver{
		mDnsServerServiceType: opts.MDnsServerServiceType,
		mDnsServerDomain:      opts.MDnsServerDomain,
	}
}

func (r *NodeNameResolver) ResolveToMDnsUrl(nodeName domain.NodeName) string {
	return fmt.Sprintf("%s.%s.%s", r.mDnsServerServiceType, nodeName, r.mDnsServerDomain)
}
