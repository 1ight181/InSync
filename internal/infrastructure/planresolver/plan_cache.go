package sync

import (
	"errors"
	"insync/internal/domain"
)

type ISyncPlanCache interface {
	GetPlan(rootName domain.RootName) (domain.SyncPlan, error)
	SetPlan(rootName domain.RootName, plan domain.SyncPlan)
}

var (
	ErrPlanNotFound = errors.New("plan cache not found")
)

type SyncPlanCache struct {
	planMap map[domain.RootName]domain.SyncPlan
}

func NewSyncPlanCache() *SyncPlanCache {
	return &SyncPlanCache{
		planMap: make(map[domain.RootName]domain.SyncPlan),
	}
}

func (s *SyncPlanCache) GetPlan(rootName domain.RootName) (domain.SyncPlan, error) {
	if plan, ok := s.planMap[rootName]; ok {
		return plan, nil
	}

	return domain.SyncPlan{}, ErrPlanNotFound
}

func (s *SyncPlanCache) SetPlan(rootName domain.RootName, plan domain.SyncPlan) {
	s.planMap[rootName] = plan
}
