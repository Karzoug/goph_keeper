package client

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rs/xid"

	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	pb "github.com/Karzoug/goph_keeper/common/grpc"
)

func (c *Client) ListVaultItemsIDName(ctx context.Context) ([]vault.IDName, error) {
	const op = "list vault items names"

	if !c.HasLocalCredintials() {
		return nil, ErrUserNeedAuthentication
	}

	names, err := c.storage.ListVaultItemsIDName(ctx)
	if err != nil {
		c.logger.Debug(op, err)
		return nil, ErrAppInternal
	}

	return names, nil
}

func (c *Client) EncryptAndSetVaultItem(ctx context.Context, item vault.Item, value any) error {
	const op = "client: encrypt and set vault item"

	if !c.HasLocalCredintials() {
		return ErrUserNeedAuthentication
	}

	if len(item.ID) == 0 {
		item.ID = xid.New().String()
	}

	if err := item.EncryptAndSetValue(value, c.credentials.encrKey); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	if err := c.storage.SetVaultItem(ctx, item); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	ctx, err := c.newContextWithAuthData(ctx)
	if err != nil {
		return nil
	}

	resp, err := c.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
		Item: &pb.VaultItem{
			Id:              item.ID,
			Name:            item.Name,
			Itype:           pb.IType(item.Type),
			Value:           item.Value,
			ClientUpdatedAt: item.ClientUpdatedAt,
			ServerUpdatedAt: item.ServerUpdatedAt,
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, pb.ErrEmptyAuthData),
			errors.Is(err, pb.ErrEmptyAuthData),
			errors.Is(err, pb.ErrUserInvalidToken),
			errors.Is(err, pb.ErrUserNeedAuthentication):
			_ = c.clearToken()
			return nil
		case errors.Is(err, pb.ErrVaultItemConflictVersion):
			// it's ok, next update method iteration hadle this conflict
			return ErrConflictVersion
		default:
			c.logger.Debug(op, err)
			if status.Code(err) == codes.Unavailable {
				return ErrServerUnavailable
			}
			return ErrServerInternal
		}
	}

	item.ServerUpdatedAt = resp.ServerUpdatedAt
	if err := c.storage.SetVaultItem(ctx, item); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	return nil
}

func (c *Client) DecryptAndGetVaultItem(ctx context.Context, id string) (vault.Item, any, error) {
	const op = "client: decrypt and get vault item"

	if !c.HasLocalCredintials() {
		return vault.Item{}, nil, ErrUserNeedAuthentication
	}

	item, err := c.storage.GetVaultItem(ctx, id)
	if err != nil {
		c.logger.Debug(op, err)
		return vault.Item{}, nil, ErrAppInternal
	}

	value, err := item.DecryptAnGetValue(c.credentials.encrKey)
	if err != nil {
		c.logger.Debug(op, err)
		return vault.Item{}, nil, ErrAppInternal
	}
	item.Value = nil

	return item, value, nil
}
