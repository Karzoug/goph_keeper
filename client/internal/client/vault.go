package client

import (
	"context"
	"errors"

	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	pb "github.com/Karzoug/goph_keeper/common/grpc"
	cvault "github.com/Karzoug/goph_keeper/common/model/vault"
)

func (c *Client) UpdateVaultItems(ctx context.Context) error {
	const op = "client: update vault items"

	ctx, err := c.newContextWithAuthData(ctx)
	if err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	// TODO: add since last update
	resp, err := c.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{})
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

	items := make([]vault.Item, len(resp.Items))
	for i := 0; i < len(resp.Items); i++ {
		items[i] = vault.Item{
			Name:      resp.Items[i].Name,
			Type:      cvault.ItemType(resp.Items[i].Itype),
			Value:     resp.Items[i].Value,
			UpdatedAt: resp.Items[i].UpdatedAt.AsTime(),
		}
	}

	if err := c.storage.SetVaultItems(items); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	return nil
}

func (c *Client) ListVaultItemsNames(ctx context.Context) ([]string, error) {
	const op = "client: list vault items names"

	names, err := c.storage.ListVaultItemsNames()
	if err != nil {
		c.logger.Debug(op, err)
		return nil, ErrAppInternal
	}

	return names, nil
}
