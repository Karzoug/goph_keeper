package item

import (
	"context"
	"errors"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/view/common"
)

var ErrWrongItemType = errors.New("got wrong item type")

func Get(c *client.Client, id string) (vault.Item, any, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), common.StandartTimeout)
	defer cancel()

	item, value, err := c.DecryptAndGetVaultItem(ctx, id)
	if err != nil {
		return vault.Item{}, nil, err
	}

	return item, value, nil
}

func Set(c *client.Client, item vault.Item, value any) error {
	ctx, cancel := context.WithTimeout(context.TODO(), common.StandartTimeout)
	defer cancel()

	err := c.EncryptAndSetVaultItem(ctx, item, value)
	if err != nil {
		return err
	}
	return nil
}
