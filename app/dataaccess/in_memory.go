package dataaccess

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v2"
)

// InMemory implements the DataAccces interface using bigcache.
type InMemory struct {
	cache *bigcache.BigCache
}

func NewInMemory(ttl, maxEntrySize, hardMaxCacheSize int, verbose bool) (*InMemory, error) {
	config := bigcache.Config{
		Shards:             1 << 10, // 1024
		LifeWindow:         time.Duration(ttl) * time.Second,
		CleanWindow:        5 * time.Minute,
		MaxEntriesInWindow: 50,
		MaxEntrySize:       maxEntrySize,
		Verbose:            verbose,
		HardMaxCacheSize:   hardMaxCacheSize,
		OnRemove:           nil,
		OnRemoveWithReason: nil,
	}
	cache, err := bigcache.NewBigCache(config)
	if err != nil {
		return nil, err
	}
	return &InMemory{
		cache: cache,
	}, nil
}

func (mem InMemory) GetNote(ctx context.Context, key string) (data []byte, found bool, err error) {
	entry, err := mem.cache.Get("NOTE:" + key)
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return entry, true, nil
}

func (mem InMemory) GetVersion(ctx context.Context, key string) (version string, err error) {
	entry, err := mem.cache.Get("VERS:" + key)
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return "", nil
		}
		return "", err
	}
	return string(entry), nil
}

func (mem InMemory) SetNote(ctx context.Context, key string, data []byte) (version string, err error) {
	t := time.Now().Format(time.RFC3339Nano)
	if err = mem.cache.Set("VERS:"+key, []byte(t)); err != nil {
		return "", err
	}
	if err = mem.cache.Set("NOTE:"+key, data); err != nil {
		return "", err
	}
	return t, nil
}
