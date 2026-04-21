package file

import (
	"context"
	"errors"
	"insync/internal/domain"
	"io"
)

type FileUseCase struct {
	fileManager IFileManager
}

type FileUseCaseOptions struct {
	FileManager IFileManager
}

var (
	ErrInvalidOpts = errors.New("Все поля FileUseCaseOptions должны быть заполнены")
)

func NewFileUseCase(opts FileUseCaseOptions) (*FileUseCase, error) {
	if opts.FileManager == nil {
		return nil, ErrInvalidOpts
	}
	return &FileUseCase{fileManager: opts.FileManager}, nil
}

func (f *FileUseCase) GetSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	return f.fileManager.GetSnapshot(ctx, rootName)
}

func (f *FileUseCase) DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) error {
	return f.fileManager.DeleteFile(ctx, rootName, relativePath)
}

func (f *FileUseCase) PutFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path, file io.Reader) error {
	return f.fileManager.PutFile(ctx, rootName, relativePath, file)
}

func (f *FileUseCase) RenameFile(ctx context.Context, rootName domain.RootName, oldPath, newPath domain.Path) error {
	return f.fileManager.RenameFile(ctx, rootName, oldPath, newPath)
}

func (f *FileUseCase) GetFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) (io.ReadCloser, error) {
	return f.fileManager.GetFile(ctx, rootName, relativePath)
}
