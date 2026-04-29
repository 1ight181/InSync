package alias

type IAliasRepository interface {
	GetAliases() ([]string, error)
	SetAlias(newAlias string, nodeName string) error
	RemoveAlias(aliasName string) error
}
