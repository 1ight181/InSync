package sync

import (
	"testing"

	"insync/internal/domain"

	"github.com/stretchr/testify/require"
)

func TestSyncPlanCache_GetPlanNotFound_ReturnsError(t *testing.T) {
	cache := NewSyncPlanCache()

	_, err := cache.GetPlan(domain.RootName("missing"))
	require.ErrorIs(t, err, ErrPlanNotFound)
}

func TestSyncPlanCache_SetPlan_CanRetrieveStoredPlan(t *testing.T) {
	cache := NewSyncPlanCache()
	expectedPlan := domain.NewSyncPlan(nil, nil, nil)

	cache.SetPlan(domain.RootName("root"), expectedPlan)

	plan, err := cache.GetPlan(domain.RootName("root"))
	require.NoError(t, err)
	require.Equal(t, expectedPlan, plan)
}
