package usecase

import (
	"context"
	"insync/internal/domain"
	ifaces "insync/internal/interfaces"
	"io"
)

type FileUseCase struct {
	fileFinder ifaces.IFileFinder
}

type FileUseCaseOptions struct {
	FileFinder ifaces.IFileFinder
}

func NewFileUseCase(opts FileUseCaseOptions) ifaces.IFileUseCase {
	return &FileUseCase{fileFinder: opts.FileFinder}
}

func (f *FileUseCase) GetFileList(ctx context.Context, rootName domain.RootName) ([]domain.FileMetadata, error) {
	return nil, nil
}

func (f *FileUseCase) DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath) error {
	return nil
}

func (f *FileUseCase) PutFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath, file io.Reader) error {
	return nil
}

func (f *FileUseCase) RenameFile(ctx context.Context, rootName domain.RootName, oldRelativePath, newRelativePath domain.RelativePath) error {
	return nil
}

func (f *FileUseCase) GetFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath) (io.ReadCloser, error) {
	return nil, nil
}
