package nodename

import (
	"errors"
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

var (
	ErrInvalidMDnsUrlResolverOptions = errors.New("Все поля MDnsUrlResolverOptions должны быть заполнены")
)

func NewMDnsUrlResolver(opts MDnsUrlResolverOptions) (*MDnsUrlResolver, error) {
	if opts.MDnsServerDomain == "" || opts.MDnsServerServiceType == "" {
		return nil, ErrInvalidMDnsUrlResolverOptions
	}
	return &MDnsUrlResolver{
		mDnsServerServiceType: opts.MDnsServerServiceType,
		mDnsServerDomain:      opts.MDnsServerDomain,
	}, nil
}

func (r *MDnsUrlResolver) Resolve(nodeName domain.NodeName) string {
	return fmt.Sprintf("%s.%s.%s", r.mDnsServerServiceType, nodeName, r.mDnsServerDomain)
}
