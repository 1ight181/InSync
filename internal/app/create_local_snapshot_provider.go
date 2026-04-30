package app

import (
	"insync/internal/infrastructure/filemanager"
	local "insync/internal/infrastructure/local"
)

func createLocalSnapshotProvider(
	fileManager *filemanager.FileManager,
) (*local.LocalSnapshotProvider, error) {

	localSnapshotProvider, err := local.NewLocalSnapshotProvider(fileManager)
	if err != nil {
		return nil, err
	}

	return localSnapshotProvider, nil
}
