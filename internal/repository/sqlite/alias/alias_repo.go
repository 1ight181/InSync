package alias

import (
	"errors"
	"insync/internal/domain"

	"gorm.io/gorm"
)

type AliasRepository struct {
	db *gorm.DB
}

type AliasRepositoryOptions struct {
	Db *gorm.DB
}

var (
	ErrInvalidAliasRepositoryOptions = errors.New("все поля AliasRepositoryOptions должны быть заполнены")
)

func NewAliasRepository(opts AliasRepositoryOptions) (*AliasRepository, error) {
	if opts.Db == nil {
		return nil, ErrInvalidAliasRepositoryOptions
	}

	return &AliasRepository{db: opts.Db}, nil
}

func (a *AliasRepository) GetAliases() ([]string, error) {
	var aliases []Alias
	err := a.db.Find(&aliases).Error
	if err != nil {
		return nil, err
	}

	var aliasStrings []string
	for _, alias := range aliases {
		aliasStrings = append(aliasStrings, alias.Name)
	}

	return aliasStrings, nil
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

	return a.db.Save(&alias).Error
}

func (a *AliasRepository) RemoveAlias(aliasName string) error {
	return a.db.Delete(&Alias{}, "name = ?", aliasName).Error
}
