package alias

import (
	"errors"
	"insync/internal/domain"
)

type AliasProvider struct {
	AliasRepository IAliasRepository
	aliasesCache    map[domain.NodeName]string
}

type AliasProviderOptions struct {
	AliasRepository IAliasRepository
}

var (
	ErrInvalidAliasProviderOptions = errors.New("Все поля AliasProviderOptions должны быть заполнены")
)

func NewAliasProvider(opts AliasProviderOptions) (*AliasProvider, error) {
	if opts.AliasRepository == nil {
		return nil, ErrInvalidAliasProviderOptions
	}

	aliases, err := opts.AliasRepository.GetAliases()
	if err != nil {
		return nil, err
	}

	return &AliasProvider{
		AliasRepository: opts.AliasRepository,
		aliasesCache:    aliases,
	}, nil
}

func (a *AliasProvider) GetAliases() map[domain.NodeName]string {
	return a.aliasesCache
}

func (a *AliasProvider) SetAlias(newAlias string, nodeName domain.NodeName) error {
	if err := a.AliasRepository.SetAlias(newAlias, nodeName); err != nil {
		return err
	}

	a.aliasesCache[nodeName] = newAlias
	return nil
}

func (a *AliasProvider) RemoveAlias(aliasName string) error {
	if err := a.AliasRepository.RemoveAlias(aliasName); err != nil {
		return err
	}

	delete(a.aliasesCache, domain.NodeName(aliasName))
	return nil
}

func (a *AliasProvider) GetAlias(nodeName domain.NodeName) string {
	return a.aliasesCache[nodeName]
}
