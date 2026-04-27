package deviceid

import (
	"errors"
	"insync/internal/domain"
	"strings"
)

var (
	ErrinvalidNodeName = errors.New("invalid node name")
)

type RemoteDeviceIdResolver struct {
	nodeNameProvider INodeNameProvider
}

type RemoteDeviceIdResolverOptions struct {
	NodeNameProvider INodeNameProvider
}

var (
	ErrRemoteDeviceIdResolverInvalidOpts = errors.New("Все поля RemoteDeviceIdResolverOptions должны быть заполнены")
)

func NewRemoteDeviceIdResolver(opts RemoteDeviceIdResolverOptions) (*RemoteDeviceIdResolver, error) {
	if opts.NodeNameProvider == nil {
		return nil, ErrRemoteDeviceIdResolverInvalidOpts
	}
	return &RemoteDeviceIdResolver{
		nodeNameProvider: opts.NodeNameProvider,
	}, nil
}

func (r *RemoteDeviceIdResolver) Resolve() (domain.DeviceId, error) {
	nodeName, err := r.nodeNameProvider.CurrentNodeName()
	if err != nil {
		return "", err
	}

	remoteDeviceId := strings.Split(nodeName.String(), ".")[0]

	return domain.DeviceId(remoteDeviceId), nil
}
