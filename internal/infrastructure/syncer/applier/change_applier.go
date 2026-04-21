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
	case domain.Create, domain.Modify:
		content, err := a.clientFactory.CurrentClient().GetFile(ctx, rootName, change.NewRelativePath)
		if err != nil {
			return err
		}

		return a.fileManager.PutFile(ctx, rootName, change.NewRelativePath, content)

	case domain.Delete:
		return a.fileManager.DeleteFile(ctx, rootName, change.OldRelativePath)

	case domain.Move, domain.Rename:
		return a.fileManager.RenameFile(ctx, rootName, change.OldRelativePath, change.NewRelativePath)
	default:
		return ErrUnknownChangeType
	}

}

func (a *ChangeApplier) ApplyRemote(ctx context.Context, rootName domain.RootName, change domain.RemoteChange) error {
	client := a.clientFactory.CurrentClient()
	switch change.ChangeType {
	case domain.Create, domain.Modify:
		content, err := a.fileManager.GetFile(ctx, rootName, change.NewRelativePath)
		if err != nil {
			return err
		}
		return client.PutFile(ctx, content, rootName, change.NewRelativePath)
	case domain.Delete:
		return client.DeleteFile(ctx, rootName, change.OldRelativePath)
	case domain.Move, domain.Rename:
		return client.RenameFile(ctx, rootName, change.OldRelativePath, change.NewRelativePath)
	default:
		return ErrUnknownChangeType
	}
}
