package scan

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type ScanUseCase struct {
	planResolver IPlanResolver
}

type ScanUseCaseOptions struct {
	PlanResolver IPlanResolver
}

var (
	ErrInvalidScanUseCaseOptions = errors.New("Все поля ScanUseCaseOptions должны быть заполнены")
)

func NewScanUseCase(opts ScanUseCaseOptions) (*ScanUseCase, error) {
	if opts.PlanResolver == nil {
		return nil, ErrInvalidScanUseCaseOptions
	}
	return &ScanUseCase{
		planResolver: opts.PlanResolver,
	}, nil
}

func (s *ScanUseCase) PlanSyncChanges(ctx context.Context, rootName domain.RootName) (domain.SyncPlan, error) {
	return s.planResolver.Resolve(ctx, rootName, false)
}
