package alias

import (
	"errors"
	"insync/internal/domain"

	"gorm.io/gorm"
)

type AliasRepository struct {
	db *gorm.DB
}

var (
	ErrInvalidAliasRepositoryOptions = errors.New("все поля AliasRepositoryOptions должны быть заполнены")
)

func NewAliasRepository(db *gorm.DB) (*AliasRepository, error) {
	if db == nil {
		return nil, ErrInvalidAliasRepositoryOptions
	}

	return &AliasRepository{db: db}, nil
}

func (a *AliasRepository) GetAliases() (map[domain.NodeName]string, error) {
	var aliases []Alias
	err := a.db.Find(&aliases).Error
	if err != nil {
		return nil, err
	}

	aliasMap := make(map[domain.NodeName]string, len(aliases))
	for _, alias := range aliases {
		aliasMap[domain.NodeName(alias.NodeName)] = alias.Name
	}

	return aliasMap, nil
}

func (a *AliasRepository) SetAlias(newAlias string, nodeName domain.NodeName) error {
	var alias Alias
	if err := a.db.First(&alias, "node_name = ?", nodeName).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if alias.Name == newAlias {
		return nil
	}

	alias.Name = newAlias
	alias.NodeName = nodeName.String()

	return a.db.Save(&alias).Error
}

func (a *AliasRepository) RemoveAlias(aliasName string) error {
	return a.db.Delete(&Alias{}, "name = ?", aliasName).Error
}
