package app

import (
	"insync/internal/infrastructure/filemanager"
	"insync/internal/infrastructure/filemanager/hash"
	hashcache "insync/internal/infrastructure/filemanager/hash/cache"
	"insync/internal/infrastructure/filemanager/pathtree"
	"insync/internal/infrastructure/filesys"
	"insync/internal/infrastructure/root"
	hashrepo "insync/internal/repository/sqlite/hash"
	"log/slog"

	"gorm.io/gorm"
)

func createFileManager(
	hashManagerLogger *slog.Logger,
	fileManagerLogger *slog.Logger,
	db *gorm.DB,
	fileSystem *filesys.FileSystem,
	rootResolver *root.RootResolver,
	hashCacheEntryExpireUnixTime int64,
) (*filemanager.FileManager, error) {
	hashCacheRepoOpts := hashrepo.HashRepositoryOptions{
		Db:                           db,
		HashCacheEntryExpireUnixTime: hashCacheEntryExpireUnixTime,
	}

	hashCacheRepo, err := hashrepo.NewHashRepository(hashCacheRepoOpts)
	if err != nil {
		return nil, err
	}

	hashCache, err := hashcache.NewHashCache(hashCacheRepo)
	if err != nil {
		return nil, err
	}

	hashCalc := hash.NewHashCalculator()
	pathTree := pathtree.NewPathTree()

	hashManagerOpts := hash.HashManagerOptions{
		HashCache:            hashCache,
		HashCalculator:       hashCalc,
		PathTreeReader:       pathTree,
		DirtyPathsRepository: hashCacheRepo,

		Logger: hashManagerLogger,
	}

	hashManager, err := hash.NewHashManager(hashManagerOpts)
	if err != nil {
		return nil, err
	}

	fileManagerOpts := filemanager.FileManagerOptions{
		RootResolver:   rootResolver,
		HashManager:    hashManager,
		FileSystem:     fileSystem,
		PathTreeWriter: pathTree,

		Logger: fileManagerLogger,
	}

	fileManager, err := filemanager.NewFileManager(fileManagerOpts)
	if err != nil {
		return nil, err
	}

	return fileManager, nil
}
