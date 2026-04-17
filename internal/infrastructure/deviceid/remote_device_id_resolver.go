package deviceid

import (
	"errors"
	"fmt"
	"insync/internal/domain"
	"strings"
)

type IRemoteDeviceIdResolver interface {
	Resolve(nodeName domain.NodeName) (domain.DeviceId, error)
}

var (
	ErrinvalidNodeName = errors.New("invalid node name")
)

type RemoteDeviceIdResolver struct {
	mDnsServerInstanceNamePostfix string
}

type RemoteDeviceIdResolverOptions struct {
	MDnsServerInstanceNamePostfix string
}

func NewRemoteDeviceIdResolver(opts RemoteDeviceIdResolverOptions) *RemoteDeviceIdResolver {
	if opts.MDnsServerInstanceNamePostfix == "" {
		panic("Все поля RemoteDeviceIdResolverOptions должны быть заполнены")
	}
	return &RemoteDeviceIdResolver{
		mDnsServerInstanceNamePostfix: opts.MDnsServerInstanceNamePostfix,
	}
}

func (r *RemoteDeviceIdResolver) Resolve(nodeName domain.NodeName) (domain.DeviceId, error) {
	mDnsServerInstanceNamePostfixWithDot := fmt.Sprintf(".%s", r.mDnsServerInstanceNamePostfix)

	instanceNamePrefix, found := strings.CutSuffix(nodeName.String(), mDnsServerInstanceNamePostfixWithDot)
	if !found || instanceNamePrefix == "" {
		return "", ErrinvalidNodeName
	}

	return domain.DeviceId(instanceNamePrefix), nil
}
