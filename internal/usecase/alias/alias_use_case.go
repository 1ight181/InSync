package alias

import (
	"errors"
	"insync/internal/domain"
)

type AliasUseCase struct {
	aliasRepo  IAliasRepository
	aliasCache map[string]domain.NodeName
}

type AliasUseCaseOptions struct {
	AliasRepo IAliasRepository
}

var (
	ErrInvalidAliasUseCaseOptions = errors.New("все поля AliasUseCaseOptions должны быть заполнены")
)

func NewAliasUseCase(opts AliasUseCaseOptions) (*AliasUseCase, error) {
	if opts.AliasRepo == nil {
		return nil, ErrInvalidAliasUseCaseOptions
	}

	aliases, err := opts.AliasRepo.GetAliases()
	if err != nil {
		return nil, err
	}

	aliasCache := make(map[string]domain.NodeName, len(aliases))
	for _, alias := range aliases {
		aliasCache[alias] = domain.NodeName(alias)
	}

	return &AliasUseCase{
		aliasRepo:  opts.AliasRepo,
		aliasCache: aliasCache,
	}, nil
}

func (a *AliasUseCase) SetAlias(newAlias string, nodeName string) error {
	return a.aliasRepo.SetAlias(newAlias, nodeName)
}

func (a *AliasUseCase) RemoveAlias(aliasName string) error {
	return a.aliasRepo.RemoveAlias(aliasName)
}

func (a *AliasUseCase) GetAliases() ([]string, error) {
	return a.aliasRepo.GetAliases()
}
