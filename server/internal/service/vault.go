package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Karzoug/goph_keeper/common/model/vault"
	"github.com/Karzoug/goph_keeper/pkg/e"
	"github.com/Karzoug/goph_keeper/pkg/logger/slog/sl"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage"
)

const lastUpdateCacheTTL = 24 * time.Hour

func (s *Service) SetVaultItem(ctx context.Context, email string, item vault.Item) (int64, error) {
	const op = "service: set vault item"

	if len(item.Value) > int(s.cfg.StorageMaxSizeItemValue) {
		return 0, e.Wrap(op, ErrVaultItemValueTooBig)
	}

	item.ClientUpdatedAt = time.Now().UnixMicro()

	err := s.storage.SetVaultItem(ctx, email, item)
	if err != nil {
		if errors.Is(err, storage.ErrNoRecordsAffected) {
			return 0, e.Wrap(op, ErrVaultItemVersionConflict)
		}
		return 0, e.Wrap(op, err)
	}
	return item.ClientUpdatedAt, nil
}

func (s *Service) ListVaultItems(ctx context.Context, email string, since int64) ([]vault.Item, error) {
	const op = "service: list vault items"

	// first try to find in cache if there is since date
	if since != 0 {
		str, err := s.caches.lastUpdate.Get(ctx, email)
		if err != nil {
			if !errors.Is(err, storage.ErrRecordNotFound) {
				s.logger.Error(op, sl.Error(err))
			}
		} else {
			t, err := strconv.ParseInt(str, 10, 64)
			if err == nil {
				// if time in cache (on server) is older or equal than given since date
				// then return empty slice of items
				if since >= t {
					return make([]vault.Item, 0), nil
				}
			}
		}
	}

	items, err := s.storage.ListVaultItems(ctx, email, since)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	var oTime int64
	for i := 0; i < len(items); i++ {
		if items[i].ServerUpdatedAt > oTime {
			oTime = items[i].ServerUpdatedAt
		}
	}

	if oTime == 0 {
		return items, nil
	}
	err = s.caches.lastUpdate.Set(ctx, email, strconv.FormatInt(oTime, 10), lastUpdateCacheTTL)
	if err != nil {
		s.logger.Warn(op, e.Wrap(op, err))
	}
	return items, nil
}
