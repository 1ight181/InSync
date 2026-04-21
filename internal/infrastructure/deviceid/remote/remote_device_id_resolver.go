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
	nodeNameProvider              INodeNameProvider
}

type RemoteDeviceIdResolverOptions struct {
	MDnsServerInstanceNamePostfix string
	NodeNameProvider              INodeNameProvider
}

var (
	ErrRemoteDeviceIdResolverInvalidOpts = errors.New("Все поля RemoteDeviceIdResolverOptions должны быть заполнены")
)

func NewRemoteDeviceIdResolver(opts RemoteDeviceIdResolverOptions) (*RemoteDeviceIdResolver, error) {
	if opts.MDnsServerInstanceNamePostfix == "" || opts.NodeNameProvider == nil {
		return nil, ErrRemoteDeviceIdResolverInvalidOpts
	}
	return &RemoteDeviceIdResolver{
		mDnsServerInstanceNamePostfix: opts.MDnsServerInstanceNamePostfix,
	}, nil
}

func (r *RemoteDeviceIdResolver) Resolve() (domain.DeviceId, error) {
	nodeName := r.nodeNameProvider.CurrentNodeName()

	mDnsServerInstanceNamePostfixWithDot := fmt.Sprintf(".%s", r.mDnsServerInstanceNamePostfix)

	instanceNamePrefix, found := strings.CutSuffix(nodeName.String(), mDnsServerInstanceNamePostfixWithDot)
	if !found || instanceNamePrefix == "" {
		return "", ErrinvalidNodeName
	}

	return domain.DeviceId(instanceNamePrefix), nil
}
