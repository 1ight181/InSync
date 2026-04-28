package sync

import (
	"context"
	"errors"
	"testing"

	"insync/internal/domain"

	"github.com/stretchr/testify/require"
)

type stubBaseSnapshotProvider struct {
	called   bool
	rootName domain.RootName
	snap     domain.Snapshot
	err      error
}

func (s *stubBaseSnapshotProvider) GetBaseSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	s.called = true
	s.rootName = rootName
	return s.snap, s.err
}

type stubLocalSnapshotProvider struct {
	called   bool
	rootName domain.RootName
	snap     domain.Snapshot
	err      error
}

func (s *stubLocalSnapshotProvider) GetLocalSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	s.called = true
	s.rootName = rootName
	return s.snap, s.err
}

type stubRemoteSnapshotProvider struct {
	called   bool
	rootName domain.RootName
	snap     domain.Snapshot
	err      error
}

func (s *stubRemoteSnapshotProvider) GetRemoteSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	s.called = true
	s.rootName = rootName
	return s.snap, s.err
}

type stubChangesPlanner struct {
	called         bool
	baseSnapshot   domain.Snapshot
	localSnapshot  domain.Snapshot
	remoteSnapshot domain.Snapshot
	plan           domain.SyncPlan
	err            error
}

func (s *stubChangesPlanner) Plan(ctx context.Context, baseSnapshot, localSnapshot, remoteSnapshot domain.Snapshot) (domain.SyncPlan, error) {
	s.called = true
	s.baseSnapshot = baseSnapshot
	s.localSnapshot = localSnapshot
	s.remoteSnapshot = remoteSnapshot
	return s.plan, s.err
}

type stubSyncPlanCache struct {
	called     bool
	rootName   domain.RootName
	plan       domain.SyncPlan
	hasPlan    bool
	setCalled  bool
	storedPlan domain.SyncPlan
}

func (s *stubSyncPlanCache) GetPlan(rootName domain.RootName) (domain.SyncPlan, error) {
	s.called = true
	s.rootName = rootName
	if s.hasPlan {
		return s.plan, nil
	}

	return domain.SyncPlan{}, ErrPlanNotFound
}

func (s *stubSyncPlanCache) SetPlan(rootName domain.RootName, plan domain.SyncPlan) {
	s.setCalled = true
	s.rootName = rootName
	s.storedPlan = plan
}

func TestNewPlanResolver_Success(t *testing.T) {
	resolver, err := NewPlanResolver(PlanResolverOptions{
		BaseSnapshotProvider:   &stubBaseSnapshotProvider{},
		LocalSnapshotProvider:  &stubLocalSnapshotProvider{},
		RemoteSnapshotProvider: &stubRemoteSnapshotProvider{},
		ChangesPlanner:         &stubChangesPlanner{},
		SyncPlanCache:          &stubSyncPlanCache{},
	})

	require.NoError(t, err)
	require.NotNil(t, resolver)
}

func TestNewPlanResolver_InvalidOptions_ReturnsError(t *testing.T) {
	resolver, err := NewPlanResolver(PlanResolverOptions{
		LocalSnapshotProvider:  &stubLocalSnapshotProvider{},
		RemoteSnapshotProvider: &stubRemoteSnapshotProvider{},
		ChangesPlanner:         &stubChangesPlanner{},
		SyncPlanCache:          &stubSyncPlanCache{},
	})

	require.ErrorIs(t, err, ErrInvalidOpts)
	require.Nil(t, resolver)
}

func TestResolve_UsesCacheWhenAvailable(t *testing.T) {
	cachedPlan := domain.NewSyncPlan(nil, nil, nil)
	cache := &stubSyncPlanCache{hasPlan: true, plan: cachedPlan}
	resolver, err := NewPlanResolver(PlanResolverOptions{
		BaseSnapshotProvider:   &stubBaseSnapshotProvider{err: errors.New("should not be called")},
		LocalSnapshotProvider:  &stubLocalSnapshotProvider{err: errors.New("should not be called")},
		RemoteSnapshotProvider: &stubRemoteSnapshotProvider{err: errors.New("should not be called")},
		ChangesPlanner:         &stubChangesPlanner{err: errors.New("should not be called")},
		SyncPlanCache:          cache,
	})
	require.NoError(t, err)

	plan, err := resolver.Resolve(context.Background(), domain.RootName("test-root"), true)
	require.NoError(t, err)
	require.Equal(t, cachedPlan, plan)
	require.True(t, cache.called)
}

func TestResolve_RebuildsPlanAndStoresInCache(t *testing.T) {
	expectedPlan := domain.NewSyncPlan(nil, nil, nil)
	cache := &stubSyncPlanCache{}
	baseProvider := &stubBaseSnapshotProvider{snap: domain.Snapshot{}}
	localProvider := &stubLocalSnapshotProvider{snap: domain.Snapshot{}}
	remoteProvider := &stubRemoteSnapshotProvider{snap: domain.Snapshot{}}
	planner := &stubChangesPlanner{plan: expectedPlan}

	resolver, err := NewPlanResolver(PlanResolverOptions{
		BaseSnapshotProvider:   baseProvider,
		LocalSnapshotProvider:  localProvider,
		RemoteSnapshotProvider: remoteProvider,
		ChangesPlanner:         planner,
		SyncPlanCache:          cache,
	})
	require.NoError(t, err)

	plan, err := resolver.Resolve(context.Background(), domain.RootName("test-root"), false)
	require.NoError(t, err)
	require.Equal(t, expectedPlan, plan)
	require.True(t, cache.setCalled)
	require.Equal(t, expectedPlan, cache.storedPlan)
	require.True(t, baseProvider.called)
	require.True(t, localProvider.called)
	require.True(t, remoteProvider.called)
	require.True(t, planner.called)
}

func TestResolve_ReturnsErrorWhenLocalSnapshotProviderFails(t *testing.T) {
	errExpected := errors.New("local snapshot failure")
	resolver, err := NewPlanResolver(PlanResolverOptions{
		BaseSnapshotProvider:   &stubBaseSnapshotProvider{},
		LocalSnapshotProvider:  &stubLocalSnapshotProvider{err: errExpected},
		RemoteSnapshotProvider: &stubRemoteSnapshotProvider{},
		ChangesPlanner:         &stubChangesPlanner{},
		SyncPlanCache:          &stubSyncPlanCache{},
	})
	require.NoError(t, err)

	_, err = resolver.Resolve(context.Background(), domain.RootName("test-root"), false)
	require.ErrorIs(t, err, errExpected)
}
