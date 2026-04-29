package alias

import (
	"errors"
	"insync/internal/domain"
)

type AliasUseCase struct {
	aliasProvider IAliasProvider
}

type AliasUseCaseOptions struct {
	AliasProvider IAliasProvider
}

var (
	ErrInvalidAliasUseCaseOptions = errors.New("все поля AliasUseCaseOptions должны быть заполнены")
)

func NewAliasUseCase(opts AliasUseCaseOptions) (*AliasUseCase, error) {
	if opts.AliasProvider == nil {
		return nil, ErrInvalidAliasUseCaseOptions
	}

	return &AliasUseCase{
		aliasProvider: opts.AliasProvider,
	}, nil
}

func (a *AliasUseCase) SetAlias(newAlias string, nodeName domain.NodeName) error {
	return a.aliasProvider.SetAlias(newAlias, nodeName)
}

func (a *AliasUseCase) RemoveAlias(aliasName string) error {
	return a.aliasProvider.RemoveAlias(aliasName)
}

func (a *AliasUseCase) GetAliases() map[domain.NodeName]string {
	return a.aliasProvider.GetAliases()
}
