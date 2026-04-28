package local

import (
	"context"
	"errors"
	"insync/internal/domain"
)

var (
	ErrUnknownChangeType = errors.New("неизвестный тип изменения")
)

type ChangeApplier struct {
	fileManager   IFileManager
	clientFactory IClientFactory
}

type ChangeApplierOptions struct {
	FileManager   IFileManager
	ClientFactory IClientFactory
}

var (
	ErrInvalidOpts = errors.New("Все поля ChangeApplierOptions должны быть заполнены")
)

func NewChangeApplier(opts ChangeApplierOptions) (*ChangeApplier, error) {
	if opts.FileManager == nil ||
		opts.ClientFactory == nil {
		return nil, ErrInvalidOpts
	}
	return &ChangeApplier{
		fileManager:   opts.FileManager,
		clientFactory: opts.ClientFactory,
	}, nil
}

func (a *ChangeApplier) ApplyLocal(ctx context.Context, rootName domain.RootName, change domain.LocalChange) error {
	switch change.ChangeType {
	case domain.CreateFile, domain.Modify:
		scopedPath, err := domain.NewScopedPath(rootName, change.NewRelativePath)
		if err != nil {
			return err
		}

		client, err := a.clientFactory.CurrentClient()
		if err != nil {
			return err
		}

		content, err := client.GetFile(ctx, scopedPath)
		if err != nil {
			return err
		}

		return a.fileManager.PutFile(ctx, scopedPath, content)

	case domain.CreateDir:
		scopedPath, err := domain.NewScopedPath(rootName, change.NewRelativePath)
		if err != nil {
			return err
		}

		return a.fileManager.CreateDir(ctx, scopedPath)
	case domain.Delete:
		scopedPath, err := domain.NewScopedPath(rootName, change.OldRelativePath)
		if err != nil {
			return err
		}
		return a.fileManager.DeleteFile(ctx, scopedPath)

	case domain.Move, domain.Rename:
		oldScopedPath, err := domain.NewScopedPath(rootName, change.OldRelativePath)
		if err != nil {
			return err
		}
		newScopedPath, err := domain.NewScopedPath(rootName, change.NewRelativePath)
		if err != nil {
			return err
		}
		return a.fileManager.RenameFile(ctx, oldScopedPath, newScopedPath)
	default:
		return ErrUnknownChangeType
	}

}

func (a *ChangeApplier) ApplyRemote(ctx context.Context, rootName domain.RootName, change domain.RemoteChange) error {
	client, err := a.clientFactory.CurrentClient()
	if err != nil {
		return err
	}

	switch change.ChangeType {
	case domain.CreateFile, domain.Modify:
		scopedPath, err := domain.NewScopedPath(rootName, change.NewRelativePath)
		if err != nil {
			return err
		}
		content, err := a.fileManager.GetFile(ctx, scopedPath)
		if err != nil {
			return err
		}
		return client.PutFile(ctx, content, scopedPath)
	case domain.CreateDir:
		scopedPath, err := domain.NewScopedPath(rootName, change.NewRelativePath)
		if err != nil {
			return err
		}
		return client.CreateDir(ctx, scopedPath)
	case domain.Delete:
		scopedPath, err := domain.NewScopedPath(rootName, change.OldRelativePath)
		if err != nil {
			return err
		}
		return client.DeleteFile(ctx, scopedPath)
	case domain.Move, domain.Rename:
		oldScopedPath, err := domain.NewScopedPath(rootName, change.OldRelativePath)
		if err != nil {
			return err
		}
		newScopedPath, err := domain.NewScopedPath(rootName, change.NewRelativePath)
		if err != nil {
			return err
		}
		return client.RenameFile(ctx, oldScopedPath, newScopedPath)
	default:
		return ErrUnknownChangeType
	}
}
