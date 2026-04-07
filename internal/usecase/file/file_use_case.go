package usecase

import (
	"context"
	"insync/internal/domain"
	ifaces "insync/internal/interfaces"
	"io"
)

type FileUseCase struct {
	fileManager ifaces.IFileManager
}

type FileUseCaseOptions struct {
	FileManager ifaces.IFileManager
}

func NewFileUseCase(opts FileUseCaseOptions) ifaces.IFileUseCase {
	return &FileUseCase{fileManager: opts.FileManager}
}

func (f *FileUseCase) GetFileList(ctx context.Context, rootName domain.RootName) ([]domain.FileEntry, error) {
	return f.fileManager.GetFileList(ctx, rootName)
}

func (f *FileUseCase) DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath) error {
	return f.fileManager.DeleteFile(ctx, rootName, relativePath)
}

func (f *FileUseCase) PutFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath, file io.Reader) error {
	return f.fileManager.PutFile(ctx, rootName, relativePath, file)
}

func (f *FileUseCase) RenameFile(ctx context.Context, rootName domain.RootName, oldRelativePath, newRelativePath domain.RelativePath) error {
	return f.fileManager.RenameFile(ctx, rootName, oldRelativePath, newRelativePath)
}

func (f *FileUseCase) GetFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath) (io.ReadCloser, error) {
	return f.fileManager.GetFile(ctx, rootName, relativePath)
}
