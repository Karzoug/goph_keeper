package client

import (
	"context"
	"errors"
	"time"

	"github.com/rs/xid"

	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/repository/storage"
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

func (c *Client) DeleteVaultItem(ctx context.Context, id string) error {
	const op = "delete vault item"

	if !c.HasLocalCredintials() {
		return ErrUserNeedAuthentication
	}

	item, err := c.storage.GetVaultItem(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			return nil
		}
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	item.Name = ""
	item.Value = nil
	item.IsDeleted = true
	item.ClientUpdatedAt = time.Now().UnixMicro()

	if err := c.storage.SetVaultItem(ctx, item); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	t, err := c.sendVaultItem(ctx, item)
	if err != nil {
		c.logger.Debug(op, err)
		return err
	}

	item.ServerUpdatedAt = t
	if err := c.storage.SetVaultItem(ctx, item); err != nil {
		c.logger.Debug(op, err)
		return ErrAppInternal
	}

	return nil
}

func (c *Client) EncryptAndSetVaultItem(ctx context.Context, item vault.Item, value any) error {
	const op = "client: encrypt and set vault item"

	if !c.HasLocalCredintials() {
		return ErrUserNeedAuthentication
	}

	if len(item.ID) == 0 {
		item.ID = xid.New().String()
	}
	item.ClientUpdatedAt = time.Now().UnixMicro()

	if err := item.EncryptAndSetValue(value, c.credentials.EncrKey); err != nil {
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

	t, err := c.sendVaultItem(ctx, item)
	if err != nil {
		c.logger.Debug(op, err)
		return err
	}

	item.ServerUpdatedAt = t
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

	value, err := item.DecryptAnGetValue(c.credentials.EncrKey)
	if err != nil {
		c.logger.Debug(op, err)
		return vault.Item{}, nil, ErrAppInternal
	}
	item.Value = nil

	return item, value, nil
}
