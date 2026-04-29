package alias

import (
	"errors"
	"insync/internal/domain"
)

type AliasUseCase struct {
	aliasRepository IAliasRepository
	aliasCache      map[string]domain.NodeName
}

type AliasUseCaseOptions struct {
	AliasRepository IAliasRepository
}

var (
	ErrInvalidAliasUseCaseOptions = errors.New("все поля AliasUseCaseOptions должны быть заполнены")
)

func NewAliasUseCase(opts AliasUseCaseOptions) (*AliasUseCase, error) {
	if opts.AliasRepository == nil {
		return nil, ErrInvalidAliasUseCaseOptions
	}

	aliases, err := opts.AliasRepository.GetAliases()
	if err != nil {
		return nil, err
	}

	aliasCache := make(map[string]domain.NodeName, len(aliases))
	for _, alias := range aliases {
		aliasCache[alias] = domain.NodeName(alias)
	}

	return &AliasUseCase{
		aliasRepository: opts.AliasRepository,
		aliasCache:      aliasCache,
	}, nil
}

func (a *AliasUseCase) SetAlias(newAlias string, nodeName domain.NodeName) error {
	return a.aliasRepository.SetAlias(newAlias, nodeName)
}

func (a *AliasUseCase) RemoveAlias(aliasName string) error {
	return a.aliasRepository.RemoveAlias(aliasName)
}

func (a *AliasUseCase) GetAliases() ([]string, error) {
	return a.aliasRepository.GetAliases()
}
