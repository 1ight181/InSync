package alias

import (
	"errors"
	"insync/internal/domain"
)

type AliasProvider struct {
	aliasRepository  IAliasRepository
	nodeToAliasCache map[domain.NodeName]string
	aliasToNodeCache map[string]domain.NodeName
}

var (
	ErrInvalidAliasProviderOptions = errors.New("Все поля AliasProvider не должны быть nil")
)

func NewAliasProvider(aliasRepository IAliasRepository) (*AliasProvider, error) {
	if aliasRepository == nil {
		return nil, ErrInvalidAliasProviderOptions
	}

	nodeToAliasCache, err := aliasRepository.GetAliases()
	if err != nil {
		return nil, err
	}

	aliasToNodeCache := make(map[string]domain.NodeName)
	for nodeName, alias := range nodeToAliasCache {
		aliasToNodeCache[alias] = nodeName
	}

	return &AliasProvider{
		aliasRepository:  aliasRepository,
		nodeToAliasCache: nodeToAliasCache,
		aliasToNodeCache: aliasToNodeCache,
	}, nil
}

func (a *AliasProvider) GetAliases() map[domain.NodeName]string {
	return a.nodeToAliasCache
}

func (a *AliasProvider) SetAlias(newAlias string, nodeName domain.NodeName) error {
	if err := a.aliasRepository.SetAlias(newAlias, nodeName); err != nil {
		return err
	}

	a.nodeToAliasCache[nodeName] = newAlias
	a.aliasToNodeCache[newAlias] = nodeName
	return nil
}

func (a *AliasProvider) RemoveAlias(aliasName string) error {
	if err := a.aliasRepository.RemoveAlias(aliasName); err != nil {
		return err
	}

	nodeNameToDelete := a.aliasToNodeCache[aliasName]
	delete(a.nodeToAliasCache, nodeNameToDelete)
	delete(a.aliasToNodeCache, aliasName)

	return nil
}

func (a *AliasProvider) GetAlias(nodeName domain.NodeName) string {
	return a.nodeToAliasCache[nodeName]
}

func (a *AliasProvider) GetNodeByAlias(aliasName string) (domain.NodeName, bool) {
	nodeName, ok := a.aliasToNodeCache[aliasName]
	return nodeName, ok
}

func (a *AliasProvider) GetAliasByNode(nodeName domain.NodeName) (string, bool) {
	alias, ok := a.nodeToAliasCache[nodeName]
	return alias, ok
}
