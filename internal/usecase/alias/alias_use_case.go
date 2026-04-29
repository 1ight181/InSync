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
	if err := a.aliasRepository.SetAlias(newAlias, nodeName); err != nil {
		return err
	}

	a.aliasCache[newAlias] = nodeName
	return nil
}

func (a *AliasUseCase) RemoveAlias(aliasName string) error {
	if err := a.aliasRepository.RemoveAlias(aliasName); err != nil {
		return err
	}

	delete(a.aliasCache, aliasName)
	return nil
}

func (a *AliasUseCase) GetAliases() map[string]domain.NodeName {
	return a.aliasCache
}
