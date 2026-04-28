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

func (f *FileUseCase) DeleteFile(ctx context.Context, scopedPath domain.ScopedPath) error {
	return f.fileManager.DeleteFile(ctx, scopedPath)
}

func (f *FileUseCase) PutFile(ctx context.Context, scopedPath domain.ScopedPath, file io.Reader) error {
	return f.fileManager.PutFile(ctx, scopedPath, file)
}

func (f *FileUseCase) RenameFile(ctx context.Context, scopedOldPath domain.ScopedPath, scopedNewPath domain.ScopedPath) error {
	return f.fileManager.RenameFile(ctx, scopedOldPath, scopedNewPath)
}

func (f *FileUseCase) GetFile(ctx context.Context, scopedPath domain.ScopedPath) (io.ReadCloser, error) {
	return f.fileManager.GetFile(ctx, scopedPath)
}

func (f *FileUseCase) CreateDir(ctx context.Context, scopedPath domain.ScopedPath) error {
	return f.fileManager.CreateDir(ctx, scopedPath)
}
