package sync

import (
	"context"
	"insync/internal/domain"
)

type SyncUseCase struct {
	syncPlanResolver IPlanResolver
	syncer           ISyncer
}

type SyncUseCaseOptions struct {
	SyncPlanResolver IPlanResolver
	Syncer           ISyncer
}

func NewSyncUseCase(options SyncUseCaseOptions) *SyncUseCase {
	if options.SyncPlanResolver == nil ||
		options.Syncer == nil {
		panic("Все поля SyncUseCaseOptions должны быть заполнены")
	}
	return &SyncUseCase{
		syncPlanResolver: options.SyncPlanResolver,
	}
}

func (s *SyncUseCase) ApplySyncChanges(ctx context.Context, shouldUseCache bool, rootName domain.RootName) (<-chan domain.ChangeEvent, error) {
	plan, err := s.syncPlanResolver.Resolve(ctx, rootName, shouldUseCache)
	if err != nil {
		return nil, err
	}
	return s.syncer.Sync(plan)
}
