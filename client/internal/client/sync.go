package client

import (
	"context"
	"errors"
	"sort"

	"golang.org/x/exp/slog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/repository/storage"
	pb "github.com/Karzoug/goph_keeper/common/grpc"
	cvault "github.com/Karzoug/goph_keeper/common/model/vault"
	"github.com/Karzoug/goph_keeper/pkg/logger/slog/sl"
)

// SyncVaultItems synchronizes client and server vault data.
func (c *Client) SyncVaultItems(ctx context.Context) error {
	if !c.HasLocalCredintials() {
		return ErrUserNeedAuthentication
	}
	ctx, err := c.newContextWithAuthData(ctx)
	if err != nil {
		return ErrUserNeedAuthentication
	}

	if err := c.updateVaultItemsFromServer(ctx); err != nil {
		return err
	}
	if err := c.sendModifiedVaultItemsToServer(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Client) updateVaultItemsFromServer(ctx context.Context) error {
	const op = "update vault items from server"

	// looking for the time of the last entry received from the server
	since, err := c.storage.GetLastServerUpdatedAt(ctx)
	if err != nil {
		if !errors.Is(err, storage.ErrRecordNotFound) {
			c.logger.Debug(op, err)
		}
		since = 0
	}

	// ask the server if there have been updates since then
	resp, err := c.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{
		Since: since,
	})
	if err != nil {
		switch {
		case errors.Is(err, pb.ErrEmptyAuthData),
			errors.Is(err, pb.ErrUserInvalidToken),
			errors.Is(err, pb.ErrUserNeedAuthentication):
			c.logger.Debug(op, sl.Error(err))
			_ = c.clearToken()
			return ErrUserNeedAuthentication
		default:
			c.logger.Debug(op, err)
			if status.Code(err) == codes.Unavailable {
				return ErrServerUnavailable
			}
			return ErrServerInternal
		}
	}

	// process items in chronological order -
	// this will allow us to return to the process later in case of an error and not get conflicts
	sort.Slice(resp.Items, func(i, j int) bool {
		return resp.Items[i].ServerUpdatedAt < resp.Items[j].ServerUpdatedAt
	})
	for i := 0; i < len(resp.Items); i++ {
		item := vault.Item{
			ID:              resp.Items[i].Id,
			Name:            resp.Items[i].Name,
			Type:            cvault.ItemType(resp.Items[i].Itype),
			Value:           resp.Items[i].Value,
			ServerUpdatedAt: resp.Items[i].ServerUpdatedAt,
			ClientUpdatedAt: resp.Items[i].ServerUpdatedAt,
		}
		dbItem, err := c.storage.GetVaultItem(ctx, item.ID)
		if err != nil {
			if !errors.Is(err, storage.ErrRecordNotFound) {
				c.logger.Debug(op, err)
				return ErrAppInternal
			}
		}
		// case: conflict version on server and client,
		// move client item version to conflict db table and
		// save server item version to main vault table
		if !(dbItem.ServerUpdatedAt == dbItem.ClientUpdatedAt) &&
			item.ServerUpdatedAt > dbItem.ServerUpdatedAt {
			err := c.storage.MoveVaultItemToConflict(ctx, item.ID)
			if err != nil {
				c.logger.Debug(op, err)
				return ErrAppInternal
			}
		}
		err = c.storage.SetVaultItem(ctx, item)
		if err != nil {
			c.logger.Debug(op, err)
			return ErrAppInternal
		}
	}

	return nil
}

func (c *Client) sendModifiedVaultItemsToServer(ctx context.Context) error {
	const op = "send modified vault items to server"

	// get all modified items from storage
	modifiedItems, err := c.storage.ListModifiedVaultItems(ctx)
	if err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	// process items in chronological order -
	// this will allow us to return to the process later in case of an error and not get conflicts
	sort.Slice(modifiedItems, func(i, j int) bool {
		return modifiedItems[i].ServerUpdatedAt < modifiedItems[j].ServerUpdatedAt
	})
	for i := 0; i < len(modifiedItems); i++ {
		var (
			serverTime int64
			err        error
		)
		if modifiedItems[i].Type == cvault.BinaryLarge {
			serverTime, err = c.sendLargeVaultItem(ctx, modifiedItems[i])
		} else {
			serverTime, err = c.sendVaultItem(ctx, modifiedItems[i])
		}
		if err != nil {
			switch {
			case errors.Is(err, ErrConflictVersion):
				// usually this is not happened,
				// but if so, we need to exit here,
				// next method iteration hadle this conflict
				return nil
			case errors.Is(err, ErrUserNeedAuthentication):
				_ = c.clearToken()
				return nil
			}
			return err
		}

		// if synchronization for this item was successful,
		// update item server time
		modifiedItems[i].ServerUpdatedAt = serverTime
		if err := c.storage.SetVaultItem(ctx, modifiedItems[i]); err != nil {
			c.logger.Debug(op, err)
			return ErrServerInternal
		}
	}
	return nil
}

func (c *Client) sendLargeVaultItem(ctx context.Context, item vault.Item) (int64, error) {
	// TODO: implement me
	panic("not implemented")
}

func (c *Client) sendVaultItem(ctx context.Context, item vault.Item) (int64, error) {
	const op = "send modified small vault item to server"

	resp, err := c.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
		Item: &pb.VaultItem{
			Id:              item.ID,
			Name:            item.Name,
			Itype:           pb.IType(item.Type),
			Value:           item.Value,
			ServerUpdatedAt: item.ServerUpdatedAt,
			ClientUpdatedAt: item.ClientUpdatedAt,
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, pb.ErrEmptyAuthData),
			errors.Is(err, pb.ErrEmptyAuthData),
			errors.Is(err, pb.ErrUserInvalidToken),
			errors.Is(err, pb.ErrUserNeedAuthentication):
			return 0, ErrUserNeedAuthentication
		case errors.Is(err, pb.ErrVaultItemConflictVersion):
			return 0, ErrConflictVersion
		default:
			c.logger.Debug(op, err)
			if status.Code(err) == codes.Unavailable {
				return 0, ErrServerUnavailable
			}
			return 0, ErrServerInternal
		}
	}

	return resp.ServerUpdatedAt, nil
}

func (c *Client) newContextWithAuthData(ctx context.Context) (context.Context, error) {
	if !c.HasToken() {
		return ctx, pb.ErrEmptyAuthData
	}
	md := metadata.New(map[string]string{"token": c.credentials.token})
	return metadata.NewOutgoingContext(ctx, md), nil
}
