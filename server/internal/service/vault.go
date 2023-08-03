package service

import (
	"context"
	"errors"
	"time"

	"github.com/Karzoug/goph_keeper/common/model/vault"
	"github.com/Karzoug/goph_keeper/pkg/e"
	"github.com/Karzoug/goph_keeper/pkg/logger/slog/sl"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage"
)

const lastUpdateCacheTTL = 24 * time.Hour

func (s *Service) SetVaultItem(ctx context.Context, email string, item vault.Item) (time.Time, error) {
	const op = "service: set vault item"

	if len(item.Value) > int(s.cfg.StorageMaxSizeItemValue) {
		return time.Time{}, e.Wrap(op, ErrVaultItemValueTooBig)
	}

	err := s.storage.SetVaultItem(ctx, email, item)
	if err != nil {
		if errors.Is(err, storage.ErrNoRecordsAffected) {
			return time.Time{}, e.Wrap(op, ErrVaultItemVersionConflict)
		}
		return time.Time{}, e.Wrap(op, err)
	}
	return item.ClientUpdatedAt, nil
}

func (s *Service) ListVaultItems(ctx context.Context, email string, since *time.Time) ([]vault.Item, error) {
	const op = "service: list vault items"

	var oTime time.Time

	// first try to find in cache if there is since date
	if since != nil {
		str, err := s.caches.lastUpdate.Get(ctx, email)
		if err != nil {
			if !errors.Is(err, storage.ErrRecordNotFound) {
				s.logger.Error(op, sl.Error(err))
			}
		} else {
			t, err := time.Parse(time.RFC3339, str)
			if err == nil {
				// if time in cache (on server) is older than given since date
				// then return empty slice of items and time of last update
				if !t.After(*since) {
					return make([]vault.Item, 0), nil
				}
			}
		}
	}

	items, err := s.storage.ListVaultItems(ctx, email, since)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	for i := 0; i < len(items); i++ {
		if items[i].ServerUpdatedAt.After(oTime) {
			oTime = items[i].ServerUpdatedAt
		}
	}

	err = s.caches.lastUpdate.Set(ctx, email, oTime.Format(time.RFC3339), lastUpdateCacheTTL)
	if err != nil {
		s.logger.Warn(op, e.Wrap(op, err))
	}

	return items, nil
}
