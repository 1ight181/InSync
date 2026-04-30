package alias

import (
	"errors"
	"insync/internal/domain"
)

type AliasProvider struct {
	aliasRepository IAliasRepository
	aliasesCache    map[domain.NodeName]string
}

var (
	ErrInvalidAliasProviderOptions = errors.New("Все поля AliasProvider не должны быть nil")
)

func NewAliasProvider(aliasRepository IAliasRepository) (*AliasProvider, error) {
	if aliasRepository == nil {
		return nil, ErrInvalidAliasProviderOptions
	}

	aliases, err := aliasRepository.GetAliases()
	if err != nil {
		return nil, err
	}

	return &AliasProvider{
		aliasRepository: aliasRepository,
		aliasesCache:    aliases,
	}, nil
}

func (a *AliasProvider) GetAliases() map[domain.NodeName]string {
	return a.aliasesCache
}

func (a *AliasProvider) SetAlias(newAlias string, nodeName domain.NodeName) error {
	if err := a.aliasRepository.SetAlias(newAlias, nodeName); err != nil {
		return err
	}

	a.aliasesCache[nodeName] = newAlias
	return nil
}

func (a *AliasProvider) RemoveAlias(aliasName string) error {
	if err := a.aliasRepository.RemoveAlias(aliasName); err != nil {
		return err
	}

	for nodeName := range a.aliasesCache {
		if a.aliasesCache[nodeName] == aliasName {
			delete(a.aliasesCache, nodeName)
		}
	}

	return nil
}

func (a *AliasProvider) GetAlias(nodeName domain.NodeName) string {
	return a.aliasesCache[nodeName]
}
