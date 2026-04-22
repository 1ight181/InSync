package root

import (
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

func (r *RootRepository) AddRoot(rootName domain.RootName, path domain.Path) error {
	root := r.toRoot(rootName, path)
	return r.db.Save(&root).Error
}

func (r *RootRepository) RemoveRoot(rootName domain.RootName) error {
	return r.db.Delete(&Root{}, "root_name = ?", rootName).Error
}

func (r *RootRepository) toRoot(rootName domain.RootName, path domain.Path) Root {
	return Root{RootName: rootName.String(), RootPath: path.String()}
}
