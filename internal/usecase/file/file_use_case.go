package file

import (
	"context"
	"insync/internal/domain"
	"io"
)

type FileUseCase struct {
	fileManager IFileManager
}

type FileUseCaseOptions struct {
	FileManager IFileManager
}

func NewFileUseCase(opts FileUseCaseOptions) *FileUseCase {
	return &FileUseCase{fileManager: opts.FileManager}
}

func (f *FileUseCase) GetFileList(ctx context.Context, rootName domain.RootName) ([]domain.FileEntry, error) {
	return f.fileManager.GetFileList(ctx, rootName)
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
