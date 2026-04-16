package sync

import "insync/internal/domain"

type ISyncPlanCache interface {
	GetPlan(rootName domain.RootName) (domain.SyncPlan, error)
	SetPlan(rootName domain.RootName, plan domain.SyncPlan)
}
