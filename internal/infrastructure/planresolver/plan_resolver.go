package sync

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type PlanResolver struct {
	baseSnapshotProvider   IBaseSnapshotProvider
	localSnapshotProvider  ILocalSnapshotProvider
	remoteSnapshotProvider IRemoteSnapshotProvider

	changesPlanner IChangesPlanner
	syncPlanCache  ISyncPlanCache
}

type PlanResolverOptions struct {
	BaseSnapshotProvider   IBaseSnapshotProvider
	LocalSnapshotProvider  ILocalSnapshotProvider
	RemoteSnapshotProvider IRemoteSnapshotProvider

	ChangesPlanner IChangesPlanner
	SyncPlanCache  ISyncPlanCache
}

var (
	ErrInvalidOpts = errors.New("Все поля PlanResolverOptions должны быть заполнены")
)

func NewPlanResolver(options PlanResolverOptions) (*PlanResolver, error) {
	if options.BaseSnapshotProvider == nil ||
		options.LocalSnapshotProvider == nil ||
		options.RemoteSnapshotProvider == nil ||
		options.ChangesPlanner == nil ||
		options.SyncPlanCache == nil {
		return nil, ErrInvalidOpts
	}
	return &PlanResolver{
		baseSnapshotProvider:   options.BaseSnapshotProvider,
		localSnapshotProvider:  options.LocalSnapshotProvider,
		remoteSnapshotProvider: options.RemoteSnapshotProvider,
		changesPlanner:         options.ChangesPlanner,
		syncPlanCache:          options.SyncPlanCache,
	}, nil
}

func (p *PlanResolver) Resolve(ctx context.Context, rootName domain.RootName, shouldUseCache bool) (domain.SyncPlan, error) {
	if shouldUseCache {
		plan, err := p.syncPlanCache.GetPlan(rootName)
		if err == nil {
			return plan, nil
		}
	}

	baseSnapshot, err := p.baseSnapshotProvider.GetBaseSnapshot(ctx, rootName)
	if err != nil {
		return domain.SyncPlan{}, err
	}
	localSnapshot, err := p.localSnapshotProvider.GetLocalSnapshot(ctx, rootName)
	if err != nil {
		return domain.SyncPlan{}, err
	}
	remoteSnapshot, err := p.remoteSnapshotProvider.GetRemoteSnapshot(ctx, rootName)
	if err != nil {
		return domain.SyncPlan{}, err
	}

	plan, err := p.changesPlanner.Plan(ctx, baseSnapshot, localSnapshot, remoteSnapshot)
	if err != nil {
		return domain.SyncPlan{}, err
	}

	p.syncPlanCache.SetPlan(rootName, plan)

	return plan, nil
}
