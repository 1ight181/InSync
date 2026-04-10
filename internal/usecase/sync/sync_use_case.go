package sync

type SyncUseCase struct {
	clientFabric IClientFabric
}

type SyncUseCaseOptions struct {
	ClientFabric IClientFabric
}

func NewSyncUseCase(opts SyncUseCaseOptions) *SyncUseCase {
	if opts.ClientFabric == nil {
		panic("Все поля SyncUseCaseOptions должны быть заполнены")
	}
	return &SyncUseCase{
		clientFabric: opts.ClientFabric,
	}
}
