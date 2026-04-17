package local

import "insync/internal/domain"

type LocalDeviceIdResolver struct {
	localDeviceIdCreator ILocalDeviceIdCreator
	currentDeviceId      domain.DeviceId
}

type LocalDeviceIdResolverOptions struct {
	LocalDeviceIdCreator ILocalDeviceIdCreator
}

func NewLocalDeviceIdResolver(opts LocalDeviceIdResolverOptions) *LocalDeviceIdResolver {
	if opts.LocalDeviceIdCreator == nil {
		panic("Все поля LocalDeviceIdResolverOptions должны быть заполнены")
	}
	return &LocalDeviceIdResolver{localDeviceIdCreator: opts.LocalDeviceIdCreator}
}

func (r *LocalDeviceIdResolver) GetLocalDeviceId() (domain.DeviceId, error) {
	if r.currentDeviceId == "" {
		newDeviceId, err := r.localDeviceIdCreator.CreateLocalDeviceId()
		if err != nil {
			return "", err
		}

		r.currentDeviceId = newDeviceId
	}

	return r.currentDeviceId, nil
}
