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

var (
	ErrRemoteDeviceIdResolverInvalidOpts = errors.New("Все поля RemoteDeviceIdResolverOptions должны быть заполнены")
)

func NewRemoteDeviceIdResolver(nodeNameProvider INodeNameProvider) (*RemoteDeviceIdResolver, error) {
	if nodeNameProvider == nil {
		return nil, ErrRemoteDeviceIdResolverInvalidOpts
	}
	return &RemoteDeviceIdResolver{
		nodeNameProvider: nodeNameProvider,
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
