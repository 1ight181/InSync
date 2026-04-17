package local

import "context"

type LocalDeviceIdResolver struct {
	localDeviceIdCreator ILocalDeviceIdCreator
	currentDeviceId      string
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

func (r *LocalDeviceIdResolver) GetLocalDeviceId(ctx context.Context) (string, error) {
	if r.currentDeviceId == "" {
		newDeviceId, err := r.localDeviceIdCreator.CreateLocalDeviceId(ctx)
		if err != nil {
			return "", err
		}

		r.currentDeviceId = newDeviceId
	}

	return r.currentDeviceId, nil
}
