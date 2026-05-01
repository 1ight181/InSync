package root

import (
	"context"
	"errors"
	"insync/internal/domain"

	"gorm.io/gorm"
)

type RootRepository struct {
	db *gorm.DB
}

func NewRootRepository(db *gorm.DB) *RootRepository {
	return &RootRepository{db: db}
}

func (r *RootRepository) GetRoots() (map[domain.RootName]domain.Path, error) {
	var roots []Root
	if err := r.db.Find(&roots).Error; err != nil {
		return nil, err
	}

	rootMap := make(map[domain.RootName]domain.Path)
	for _, root := range roots {
		// Валидность данных гарантируется контрактом добавляемых в БД данных
		rootMap[domain.RootName(root.RootName)] = domain.Path(root.RootPath)
	}
	return rootMap, nil
}

func (r *RootRepository) AddRoot(ctx context.Context, rootName domain.RootName, path domain.Path) error {
	var root Root
	err := r.db.First(&root, "root_name = ?", rootName).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	root.RootName = rootName.String()
	root.RootPath = path.String()

	err = r.db.WithContext(ctx).Save(&root).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrRootAlreadyExists
	}

	return err
}

func (r *RootRepository) RemoveRoot(ctx context.Context, rootName domain.RootName) error {
	return r.db.WithContext(ctx).Delete(&Root{}, "root_name = ?", rootName).Error
}

func (r *RootRepository) toRoot(rootName domain.RootName, path domain.Path) Root {
	return Root{RootName: rootName.String(), RootPath: path.String()}
}
