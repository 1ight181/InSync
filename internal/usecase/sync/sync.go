package sync

import (
	"context"
	"errors"
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

var (
	ErrInvalidSyncUseCaseOptions = errors.New("Все поля SyncUseCaseOptions должны быть заполнены")
)

func NewSyncUseCase(options SyncUseCaseOptions) (*SyncUseCase, error) {
	if options.SyncPlanResolver == nil ||
		options.Syncer == nil {
		return nil, ErrInvalidSyncUseCaseOptions
	}
	return &SyncUseCase{
		syncPlanResolver: options.SyncPlanResolver,
		syncer:           options.Syncer,
	}, nil
}

func (s *SyncUseCase) ApplySyncChanges(ctx context.Context, shouldUseCache bool, rootName domain.RootName) (
	appliedChanges <-chan domain.ChangeEvent,
	conflicts <-chan domain.Conflict,
	baseSnapshotSaveError <-chan error,
	userDecision chan<- domain.Decision,
	err error,
) {
	plan, err := s.syncPlanResolver.Resolve(ctx, rootName, shouldUseCache)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return s.syncer.Sync(ctx, plan, rootName)
}
