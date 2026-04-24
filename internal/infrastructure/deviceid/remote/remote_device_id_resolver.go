package deviceid

import (
	"errors"
	"fmt"
	"insync/internal/domain"
	"strings"
)

var (
	ErrinvalidNodeName = errors.New("invalid node name")
)

type RemoteDeviceIdResolver struct {
	mDnsServerInstanceNamePostfix string
	mDnsServerDomain              string
	nodeNameProvider              INodeNameProvider
}

type RemoteDeviceIdResolverOptions struct {
	MDnsServerInstanceNamePostfix string
	MDnsServerDomain              string
	NodeNameProvider              INodeNameProvider
}

var (
	ErrRemoteDeviceIdResolverInvalidOpts = errors.New("Все поля RemoteDeviceIdResolverOptions должны быть заполнены")
)

func NewRemoteDeviceIdResolver(opts RemoteDeviceIdResolverOptions) (*RemoteDeviceIdResolver, error) {
	if opts.MDnsServerInstanceNamePostfix == "" || opts.NodeNameProvider == nil || opts.MDnsServerDomain == "" {
		return nil, ErrRemoteDeviceIdResolverInvalidOpts
	}
	return &RemoteDeviceIdResolver{
		mDnsServerInstanceNamePostfix: opts.MDnsServerInstanceNamePostfix,
		nodeNameProvider:              opts.NodeNameProvider,
		mDnsServerDomain:              opts.MDnsServerDomain,
	}, nil
}

func (r *RemoteDeviceIdResolver) Resolve() (domain.DeviceId, error) {
	nodeName, err := r.nodeNameProvider.CurrentNodeName()
	if err != nil {
		return "", err
	}

	mDnsServerPostfixWithDotAndDomain := fmt.Sprintf(".%s.%s", r.mDnsServerInstanceNamePostfix, r.mDnsServerDomain)

	instanceNamePrefix, found := strings.CutSuffix(nodeName.String(), mDnsServerPostfixWithDotAndDomain)
	if !found || instanceNamePrefix == "" {
		return "", ErrinvalidNodeName
	}

	return domain.DeviceId(instanceNamePrefix), nil
}
