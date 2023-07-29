package smap

import (
	"context"
	"sync"
	"time"

	"github.com/Karzoug/goph_keeper/pkg/e"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage"
)

type smap struct {
	items sync.Map
	close chan struct{}
}

type item struct {
	data    string
	expires int64
}

// New creates sync map storage type of key-value.
// Simple, but thread-safe. Designed mainly for testing purposes.
func New(cleaningInterval time.Duration) *smap {
	smap := &smap{
		close: make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(cleaningInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now().UnixNano()

				smap.items.Range(func(key, value any) bool {
					item := value.(item)

					if item.expires > 0 && now > item.expires {
						smap.items.Delete(key)
					}

					return true
				})

			case <-smap.close:
				return
			}
		}
	}()

	return smap
}

// Get returns value by key.
func (smap *smap) Get(_ context.Context, key string) (string, error) {
	const op = "sync map: get"

	obj, exists := smap.items.Load(key)

	if !exists {
		return "", e.Wrap(op, storage.ErrRecordNotFound)
	}

	item := obj.(item)

	if item.expires > 0 && time.Now().UnixNano() > item.expires {
		smap.items.Delete(key)
		return "", e.Wrap(op, storage.ErrRecordNotFound)
	}

	return item.data, nil
}

// Set sets value by key.
func (smap *smap) Set(_ context.Context, key string, value string, duration time.Duration) error {
	var expires int64

	if duration > 0 {
		expires = time.Now().Add(duration).UnixNano()
	}

	smap.items.Store(key, item{
		data:    value,
		expires: expires,
	})

	return nil
}

// Range calls f sequentially for each key and value present in the storage.
func (smap *smap) Range(f func(key, value any) bool) {
	now := time.Now().UnixNano()

	fn := func(key, value any) bool {
		item := value.(item)

		if item.expires > 0 && now > item.expires {
			return true
		}

		return f(key, item.data)
	}

	smap.items.Range(fn)
}

// Delete deletes value by key.
func (smap *smap) Delete(_ context.Context, key string) error {
	smap.items.Delete(key)
	return nil
}

// Close closes storage: cleans internal map and stop cleaning work.
func (smap *smap) Close() error {
	smap.close <- struct{}{}
	smap.items = sync.Map{}

	return nil
}
