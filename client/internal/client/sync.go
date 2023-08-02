package client

import (
	"context"
	"errors"
	"sort"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/repository/storage"
	pb "github.com/Karzoug/goph_keeper/common/grpc"
	cvault "github.com/Karzoug/goph_keeper/common/model/vault"
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
	var ts *timestamppb.Timestamp
	since, err := c.storage.GetLastServerUpdatedAt(ctx)
	if err == nil {
		ts = timestamppb.New(since)
	}

	// ask the server if there have been updates since then
	resp, err := c.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{
		Since: ts,
	})
	if err != nil {
		switch {
		case errors.Is(err, pb.ErrEmptyAuthData),
			errors.Is(err, pb.ErrEmptyAuthData),
			errors.Is(err, pb.ErrUserInvalidToken),
			errors.Is(err, pb.ErrUserNeedAuthentication):
			return ErrUserNeedAuthentication
		default:
			c.logger.Debug(op, err)
			return ErrServerInternal
		}
	}

	// process items in chronological order -
	// this will allow us to return to the process later in case of an error and not get conflicts
	sort.Slice(resp.Items, func(i, j int) bool {
		return resp.Items[i].ServerUpdatedAt.Seconds < resp.Items[j].ServerUpdatedAt.Seconds
	})
	for i := 0; i < len(resp.Items); i++ {
		item := vault.Item{
			ID:              resp.Items[i].Id,
			Name:            resp.Items[i].Name,
			Type:            cvault.ItemType(resp.Items[i].Itype),
			Value:           resp.Items[i].Value,
			ServerUpdatedAt: resp.Items[i].ServerUpdatedAt.AsTime(),
			ClientUpdatedAt: resp.Items[i].ServerUpdatedAt.AsTime(),
		}
		dbItem, err := c.storage.GetVaultItem(ctx, item.Name)
		if err != nil {
			if err != storage.ErrRecordNotFound {
				c.logger.Debug(op, err)
				return ErrAppInternal
			}
		}
		// case: conflict version on server and client,
		// move client item version to conflict db table and
		// save server item version to main vault table
		if !dbItem.ServerUpdatedAt.Equal(dbItem.ClientUpdatedAt) && // dbItem not pointer so it's ok
			item.ServerUpdatedAt.After(dbItem.ServerUpdatedAt) {
			err := c.storage.MoveVaultItemToConflict(ctx, item.Name)
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
		return modifiedItems[i].ServerUpdatedAt.Unix() < modifiedItems[j].ServerUpdatedAt.Unix()
	})
	for i := 0; i < len(modifiedItems); i++ {
		resp, err := c.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:              modifiedItems[i].ID,
				Name:            modifiedItems[i].Name,
				Itype:           pb.IType(modifiedItems[i].Type),
				Value:           modifiedItems[i].Value,
				ServerUpdatedAt: timestamppb.New(modifiedItems[i].ServerUpdatedAt),
				ClientUpdatedAt: timestamppb.New(modifiedItems[i].ClientUpdatedAt),
			},
		})
		if err != nil {
			switch {
			case errors.Is(err, pb.ErrEmptyAuthData),
				errors.Is(err, pb.ErrEmptyAuthData),
				errors.Is(err, pb.ErrUserInvalidToken),
				errors.Is(err, pb.ErrUserNeedAuthentication):
				return ErrUserNeedAuthentication
			case errors.Is(err, pb.ErrVaultItemConflictVersion):
				// usually this is not happened,
				// but if so, we need to exit here,
				// next method iteration hadle this conflict
				return nil
			default:
				c.logger.Debug(op, err)
				return ErrServerInternal
			}
		}

		// if synchronization for this item was successful,
		// update item server time
		modifiedItems[i].ServerUpdatedAt = modifiedItems[i].ClientUpdatedAt
		modifiedItems[i].ID = resp.Id
		if err := c.storage.SetVaultItem(ctx, modifiedItems[i]); err != nil {
			c.logger.Debug(op, err)
			return ErrServerInternal
		}
	}
	return nil
}

func (c *Client) newContextWithAuthData(ctx context.Context) (context.Context, error) {
	if !c.HasToken() {
		return ctx, pb.ErrEmptyAuthData
	}
	md := metadata.New(map[string]string{"token": c.credentials.token})
	return metadata.NewOutgoingContext(ctx, md), nil
}
