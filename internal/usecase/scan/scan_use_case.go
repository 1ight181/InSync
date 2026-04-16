package scan

import (
	"context"
	"insync/internal/domain"
)

type ScanUseCase struct {
	planResolver IPlanResolver
}

type ScanUseCaseOptions struct {
	PlanResolver IPlanResolver
}

func NewScanUseCase(opts ScanUseCaseOptions) *ScanUseCase {
	if opts.PlanResolver == nil {
		panic("Все поля ScanUseCaseOptions должны быть заполнены")
	}
	return &ScanUseCase{
		planResolver: opts.PlanResolver,
	}
}

func (s *ScanUseCase) PlanSyncChanges(ctx context.Context, rootName domain.RootName) (domain.SyncPlan, error) {
	return s.planResolver.Resolve(ctx, rootName, false)
}
