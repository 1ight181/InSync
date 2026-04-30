package scan

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type ScanUseCase struct {
	planResolver IPlanResolver
}

var (
	ErrInvalidScanUseCaseOptions = errors.New("Все поля ScanUseCaseOptions не должны nil")
)

func NewScanUseCase(planResolver IPlanResolver) (*ScanUseCase, error) {
	if planResolver == nil {
		return nil, ErrInvalidScanUseCaseOptions
	}
	return &ScanUseCase{
		planResolver: planResolver,
	}, nil
}

func (s *ScanUseCase) PlanSyncChanges(ctx context.Context, rootName domain.RootName) (domain.SyncPlan, error) {
	return s.planResolver.Resolve(ctx, rootName, false)
}
