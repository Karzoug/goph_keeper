package item

import (
	"context"
	"errors"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/view/common"
)

var ErrWrongItemType = errors.New("got wrong item type")

func Get(ctx context.Context, c *client.Client, id string) (vault.Item, any, error) {
	ctx, cancel := context.WithTimeout(ctx, common.StandartTimeout)
	defer cancel()

	item, value, err := c.DecryptAndGetVaultItem(ctx, id)
	if err != nil {
		return vault.Item{}, nil, err
	}

	return item, value, nil
}

func Set(ctx context.Context, c *client.Client, item vault.Item, value any) error {
	ctx, cancel := context.WithTimeout(ctx, common.StandartTimeout)
	defer cancel()

	if err := c.EncryptAndSetVaultItem(ctx, item, value); err != nil {
		return err
	}
	return nil
}

func Delete(ctx context.Context, c *client.Client, id string) error {
	ctx, cancel := context.WithTimeout(ctx, common.StandartTimeout)
	defer cancel()

	if err := c.DeleteVaultItem(ctx, id); err != nil {
		return err
	}
	return nil
}
